package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lucasassuncao/gopaper/internal/models"
	"github.com/lucasassuncao/yedit/docgenerator"
	"github.com/lucasassuncao/yedit/theme"

	"github.com/spf13/cobra"
)

// ShowCmd returns the "show-docs" command, which renders documentation
// generated from the config schema directly in the terminal.
func ShowCmd() *cobra.Command {
	var themeName string
	var listThemes bool
	var section string

	cmd := &cobra.Command{
		Use:               "show-docs",
		Short:             "Show configuration reference documentation in terminal",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
		Long: `Render reference documentation for the gopaper configuration schema
directly in the terminal, generated from the same field descriptions,
defaults, and constraints shown by 'gopaper edit'.`,
		Example: `  # Browse the full reference
  gopaper show-docs

  # Jump straight to the history settings
  gopaper show-docs --section history

  # Use a different theme
  gopaper show-docs --theme dracula`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listThemes {
				names := make([]string, 0, len(theme.All()))
				for name := range theme.All() {
					names = append(names, name)
				}
				sort.Strings(names)
				for _, name := range names {
					fmt.Println(name)
				}
				return nil
			}

			t, ok := theme.All()[themeName]
			if !ok {
				return fmt.Errorf("unknown theme %q — run 'gopaper show-docs --list-themes' to see available themes", themeName)
			}

			return showDocs(t, section)
		},
	}

	cmd.Flags().StringVar(&themeName, "theme", "dark", "Theme name (run --list-themes to see options)")
	cmd.Flags().BoolVar(&listThemes, "list-themes", false, "List available theme names and exit")
	cmd.Flags().StringVar(&section, "section", "", "Show only the documentation for this topic (case-insensitive, partial match)")

	return cmd
}

func showDocs(t theme.Theme, section string) error {
	entries := []docgenerator.Entry{
		{Config: models.Configuration{}, SplitStructs: true},
		{Config: models.Categories{}, SplitStructs: true},
	}

	docs, err := docgenerator.GenerateInMemory(entries)
	if err != nil {
		return fmt.Errorf("failed to generate docs: %w", err)
	}

	if section != "" {
		filtered := filterDocSet(docs, section)
		if len(filtered.Pages) == 0 {
			available := make([]string, 0, len(docs.Pages))
			for name := range docs.Pages {
				available = append(available, strings.ToLower(name))
			}
			sort.Strings(available)
			return fmt.Errorf("no documentation found for section %q — available: %s", section, strings.Join(available, ", "))
		}
		docs = filtered
	}

	if err := docgenerator.RenderMarkdownDocsInTerminal(docs, "gopaper", t); err != nil {
		return fmt.Errorf("failed to render docs: %w", err)
	}
	return nil
}

// filterDocSet returns a new DocSet containing only pages whose name matches
// section (case-insensitive substring), plus any children of matched pages.
func filterDocSet(ds docgenerator.DocSet, section string) docgenerator.DocSet {
	q := strings.ToLower(section)
	out := docgenerator.DocSet{
		Pages:    make(map[string]string),
		Children: make(map[string][]string),
	}
	for name, page := range ds.Pages {
		if !strings.Contains(strings.ToLower(name), q) {
			continue
		}
		out.Pages[name] = page
		if children, ok := ds.Children[name]; ok {
			out.Children[name] = children
			for _, child := range children {
				if p, ok := ds.Pages[child]; ok {
					out.Pages[child] = p
				}
			}
		}
	}
	return out
}
