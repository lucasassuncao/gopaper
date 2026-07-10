package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/lucasassuncao/gopaper/internal/models"
	"github.com/lucasassuncao/yedit/editor"
	"github.com/lucasassuncao/yedit/theme"
	"github.com/spf13/cobra"
)

// EditCmd returns the "edit" command, which opens an interactive TUI editor
// for the gopaper configuration file.
func EditCmd() *cobra.Command {
	var configPath string
	var output string
	var themeName string
	var listThemes bool
	var noSaveConfirm bool
	var noDeleteConfirm bool
	var noValidateOnSave bool
	var dump bool
	var dumpPath string

	cmd := &cobra.Command{
		Use:               "edit",
		Short:             "Edit the gopaper configuration file in an interactive TUI",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
		Long: `Open the gopaper configuration file in an interactive two-panel TUI editor.

The left panel lists top-level configuration keys; pressing Enter opens the
block editor where sub-fields can be toggled and edited. Ctrl+S writes the
file; Ctrl+U undoes the last change; Ctrl+Y redoes it; Esc quits.

If --config is not provided, the editor looks for gopaper.yaml next to the
executable and in its conf subdirectory (the same locations gopaper searches
when running normally). If none is found, it falls back to creating one at
<executable_dir>/conf/gopaper.yaml.

Use --output to write to a different file than the one loaded (e.g. to
produce a new config from an existing template).`,
		Example: `  # Edit the default configuration file
  gopaper edit

  # Edit with the Dracula theme
  gopaper edit --theme dracula

  # List all available themes
  gopaper edit --list-themes

  # Edit a specific configuration file
  gopaper edit -c /path/to/gopaper.yaml

  # Load from --config but save to a new file
  gopaper edit --output /path/to/new.yaml

  # Record a session trace to attach to a bug report
  gopaper edit --dump`,
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

			selectedTheme, ok := theme.All()[themeName]
			if !ok {
				return fmt.Errorf("unknown theme %q — run 'gopaper edit --list-themes' to see available themes", themeName)
			}

			loadPath, savePath, err := resolveEditPaths(configPath, output)
			if err != nil {
				return err
			}
			output = savePath

			gopaperHints, err := buildGopaperHints()
			if err != nil {
				return fmt.Errorf("building hint source: %w", err)
			}

			res, err := editor.Run(editor.Config{
				Path:             loadPath,
				SavePath:         output,
				Schema:           &models.Config{},
				Title:            "gopaper",
				BlockPresets:     GopaperBlockPresets,
				DocPresets:       GopaperDocPresets,
				EnableHints:      true,
				Metadata:         gopaperHints,
				Theme:            selectedTheme,
				NoSaveConfirm:    noSaveConfirm,
				NoDeleteConfirm:  noDeleteConfirm,
				NoValidateOnSave: noValidateOnSave,
				Validators:       GopaperValidators,
				Dump:             dump || dumpPath != "",
				DumpPath:         dumpPath,
			})
			if err != nil {
				return err
			}
			if res.Saved {
				savedTo := loadPath
				if output != "" {
					savedTo = output
				}
				fmt.Println("configuration saved to", savedTo)
			}
			if res.DumpPath != "" {
				fmt.Println("session trace written to", res.DumpPath)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to the configuration file to edit (default: standard lookup)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Save to this file instead of the loaded config (load path is unchanged)")
	cmd.Flags().StringVar(&themeName, "theme", "dark", "Theme name (run --list-themes to see options)")
	cmd.Flags().BoolVar(&listThemes, "list-themes", false, "List available theme names and exit")
	cmd.Flags().BoolVar(&noSaveConfirm, "no-save-confirm", false, "Skip the 'Save changes?' confirmation dialog")
	cmd.Flags().BoolVar(&noDeleteConfirm, "no-delete-confirm", false, "Skip the 'Remove block?' confirmation dialog")
	cmd.Flags().BoolVar(&noValidateOnSave, "no-validate-on-save", false, "Allow saving even when validators report errors (a warning is shown)")
	cmd.Flags().BoolVar(&dump, "dump", false, "Record every editor action to a JSONL trace file for bug reports (path is printed on exit)")
	cmd.Flags().StringVar(&dumpPath, "dump-path", "", "Write the session trace to this file instead of a temp file (implies --dump)")

	return cmd
}

// resolveEditPaths decides which file the editor loads and where it saves.
// --config wins; with only --output, that file is both loaded and saved.
// Otherwise the editor searches the same locations root.go uses to find
// gopaper.yaml (the executable directory, then its conf subdirectory); only
// when no config exists in either place does it fall back to the location
// gopaper init creates.
func resolveEditPaths(configFlag, output string) (loadPath, savePath string, err error) {
	if configFlag != "" {
		return configFlag, output, nil
	}
	if output != "" {
		return output, "", nil
	}

	ex, err := os.Executable()
	if err != nil {
		return "", "", fmt.Errorf("could not determine executable path: %w", err)
	}
	exDir := filepath.Dir(ex)

	candidates := []string{
		filepath.Join(exDir, "gopaper.yaml"),
		filepath.Join(exDir, "conf", "gopaper.yaml"),
	}
	for _, candidate := range candidates {
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, "", nil
		}
	}

	return filepath.Join(exDir, "conf", "gopaper.yaml"), "", nil
}
