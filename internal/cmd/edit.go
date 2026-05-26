package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucasassuncao/gopaper/internal/models"
	"github.com/lucasassuncao/yedit/editor"
	"github.com/spf13/cobra"
)

// EditCmd returns the "edit" command, which opens an interactive TUI editor
// for the gopaper configuration file.
func EditCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:               "edit",
		Short:             "Edit the gopaper configuration file in an interactive TUI",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
		Long: `Open the gopaper configuration file in an interactive two-panel TUI editor.

The left panel lists top-level configuration keys; pressing Space opens a
tree-view where sub-fields can be toggled, edited, and saved. Ctrl+S writes
the file; Ctrl+Z undoes the last change; Esc quits.

If --config is not provided, the editor opens gopaper.yaml from the same
directory as the executable.`,
		Example: `  # Edit the default configuration file
  gopaper edit

  # Edit a specific configuration file
  gopaper edit -c /path/to/gopaper.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if configPath == "" {
				ex, err := os.Executable()
				if err != nil {
					return fmt.Errorf("could not determine executable path: %w", err)
				}
				configPath = filepath.Join(filepath.Dir(ex), "gopaper.yaml")
			}

			return editor.Run(editor.Config{
				Path:   configPath,
				Schema: &models.Config{},
				Title:  "gopaper",
			})
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to the configuration file to edit (default: <executable_dir>/gopaper.yaml)")

	return cmd
}
