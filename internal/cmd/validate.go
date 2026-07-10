package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/lucasassuncao/gopaper/internal/config"
	"github.com/lucasassuncao/gopaper/internal/models"
	"github.com/lucasassuncao/yedit/editor"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type validateFormat string

const (
	formatPretty validateFormat = "pretty"
	formatPlain  validateFormat = "plain"
	formatJSON   validateFormat = "json"
)

var validFormats = []string{string(formatPretty), string(formatPlain), string(formatJSON)}

// ValidateCmd defines the "validate" subcommand.
func ValidateCmd() *cobra.Command {
	var (
		configPath string
		format     string
		summary    bool
		strict     bool
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a configuration file and report all errors",
		// Override root's PreRunE — validate reads the file directly and must
		// not abort when the config has errors.
		PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
		Long: `Validate the gopaper configuration file without opening the TUI editor.

Runs the same validators as 'gopaper edit' (required fields, allowed values,
uniqueness, cross-field rules) and reports every violation found.`,
		Example: `  # Validate the default configuration file
  gopaper validate

  # Validate a specific file, as JSON
  gopaper validate -c /path/to/gopaper.yaml -f json

  # Also check that every category's source directory exists on disk
  gopaper validate --strict`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(configPath, validateFormat(format), summary, strict)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to the configuration file to validate (default: standard lookup)")
	cmd.Flags().StringVarP(&format, "format", "f", "pretty", fmt.Sprintf("Output format: %s", strings.Join(validFormats, ", ")))
	cmd.Flags().BoolVar(&summary, "summary", false, "Show only the error count, not individual violations")
	cmd.Flags().BoolVar(&strict, "strict", false, "Also verify that every category's source directory exists on disk")
	return cmd
}

// runValidate loads the config, runs validators, and prints results in the
// requested format. It returns an error when validation fails, so the
// process exit code reflects the result.
func runValidate(configPath string, format validateFormat, summaryOnly, strict bool) error {
	switch format {
	case formatPretty, formatPlain, formatJSON:
	default:
		return fmt.Errorf("unknown format %q — use one of: %s", format, strings.Join(validFormats, ", "))
	}

	loadPath, _, err := resolveEditPaths(configPath, "")
	if err != nil {
		return err
	}

	raw, err := os.ReadFile(loadPath) // #nosec G304 -- path comes from --config or the standard config lookup, the same trust boundary as loading the app config elsewhere
	if err != nil {
		return fmt.Errorf("could not read %s: %w", loadPath, err)
	}

	// Parse eagerly so a malformed file is reported as a hard error instead
	// of silently producing zero violations (the FromMetadata validators walk
	// a parsed YAML tree and report nothing when it fails to parse).
	var probe any
	if err := yaml.Unmarshal(raw, &probe); err != nil {
		return fmt.Errorf("could not parse %s: %w", loadPath, err)
	}

	hints, err := buildGopaperHints()
	if err != nil {
		return fmt.Errorf("building hint source: %w", err)
	}

	wired := editor.Wire(GopaperValidators, editor.Config{Schema: &models.Config{}, Metadata: hints})
	violations := editor.RunAll(wired, raw, nil)

	if strict {
		violations = append(violations, strictDirViolations(raw)...)
	}

	switch format {
	case formatJSON:
		printValidateJSON(violations, summaryOnly)
	case formatPlain:
		printValidatePlain(violations, summaryOnly)
	default:
		printValidatePretty(violations, summaryOnly)
	}

	if len(violations) > 0 {
		return errors.New("validation failed")
	}
	return nil
}

// strictDirViolations checks whether each category's source directory exists
// on disk, returning a violation for each one that doesn't.
func strictDirViolations(raw []byte) []editor.Violation {
	var doc struct {
		Categories []struct {
			Source string `yaml:"source"`
		} `yaml:"categories"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil
	}

	var out []editor.Violation
	for i, c := range doc.Categories {
		if c.Source == "" {
			continue
		}
		if _, err := os.Stat(config.ExpandTilde(c.Source)); os.IsNotExist(err) {
			out = append(out, editor.Violation{
				Path:    fmt.Sprintf("categories[%d].source", i),
				Message: fmt.Sprintf("directory does not exist: %s", c.Source),
			})
		}
	}
	return out
}

var topSectionRe = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9_-]*)`)

// sectionOf extracts the top-level section name from a violation path, or
// "(general)" if it cannot be determined.
func sectionOf(path string) string {
	if m := topSectionRe.FindString(path); m != "" {
		return m
	}
	return "(general)"
}

// groupViolations groups violations by their top-level section and returns
// sections in a stable order (alphabetical).
func groupViolations(violations []editor.Violation) ([]string, map[string][]editor.Violation) {
	bySection := make(map[string][]editor.Violation)
	for _, v := range violations {
		s := sectionOf(v.Path)
		bySection[s] = append(bySection[s], v)
	}
	sections := make([]string, 0, len(bySection))
	for s := range bySection {
		sections = append(sections, s)
	}
	sort.Strings(sections)
	return sections, bySection
}

// subPath strips the top-level section prefix from a violation path.
func subPath(path string) string {
	if i := strings.IndexAny(path, ".["); i >= 0 {
		rest := path[i:]
		return strings.TrimPrefix(rest, ".")
	}
	return path
}

// summaryLine builds the coloured summary string shared by the pretty format.
func summaryLine(sections []string, bySection map[string][]editor.Violation) string {
	parts := make([]string, 0, len(sections))
	total := 0
	for _, s := range sections {
		n := len(bySection[s])
		total += n
		parts = append(parts, fmt.Sprintf("%d in %s", n, s))
	}
	return fmt.Sprintf("%s error(s) — %s",
		pterm.Red(fmt.Sprintf("%d", total)),
		strings.Join(parts, ", "),
	)
}

// printValidatePretty renders violations grouped by section with a tree-like
// structure and a coloured summary.
func printValidatePretty(violations []editor.Violation, summaryOnly bool) {
	if len(violations) == 0 {
		pterm.Success.Println("No errors found — configuration is valid")
		return
	}

	sections, bySection := groupViolations(violations)

	if !summaryOnly {
		for _, section := range sections {
			vs := bySection[section]
			pterm.Println()
			pterm.Bold.Println("  " + section)
			for i, v := range vs {
				connector := pterm.Gray("├─")
				if i == len(vs)-1 {
					connector = pterm.Gray("└─")
				}
				sp := pterm.Yellow(fmt.Sprintf("%-30s", subPath(v.Path)))
				pterm.Printf("  %s %s %s\n", connector, sp, v.Message)
			}
		}
		pterm.Println()
	}

	pterm.Println(summaryLine(sections, bySection))
}

// printValidatePlain renders violations as plain text lines with a final count.
func printValidatePlain(violations []editor.Violation, summaryOnly bool) {
	if len(violations) == 0 {
		fmt.Println("ok")
		return
	}

	if !summaryOnly {
		for _, v := range violations {
			fmt.Printf("%-40s %s\n", v.Path, v.Message)
		}
	}
	fmt.Printf("%d error(s)\n", len(violations))
}

// printValidateJSON renders violations as a single JSON object.
func printValidateJSON(violations []editor.Violation, summaryOnly bool) {
	type jsonViolation struct {
		Path    string `json:"path"`
		Message string `json:"message"`
	}
	out := struct {
		Valid      bool            `json:"valid"`
		ErrorCount int             `json:"error_count"`
		Violations []jsonViolation `json:"violations,omitempty"`
	}{
		Valid:      len(violations) == 0,
		ErrorCount: len(violations),
	}
	if !summaryOnly {
		for _, v := range violations {
			out.Violations = append(out.Violations, jsonViolation{Path: v.Path, Message: v.Message})
		}
	}
	data, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(data))
}
