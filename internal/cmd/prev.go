package cmd

import (
	"errors"
	"fmt"

	"github.com/lucasassuncao/gopaper/internal/helper"
	"github.com/lucasassuncao/gopaper/internal/history"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// PrevCmd sets the previous wallpaper from history.
func PrevCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prev",
		Short: "Set the previous wallpaper from history",
		RunE: func(cmd *cobra.Command, args []string) error {
			histPath, err := history.DefaultPath()
			if err != nil {
				return fmt.Errorf("could not determine history path: %w", err)
			}

			h, err := history.Load(histPath)
			if err != nil {
				return fmt.Errorf("could not load history: %w", err)
			}

			entry, err := history.Prev(h)
			if err != nil {
				if errors.Is(err, history.ErrHistoryEmpty) || errors.Is(err, history.ErrAlreadyOldest) {
					pterm.Warning.Println(err)
					return nil
				}
				return err
			}

			if err := helper.SetWallpaperFromPath(entry.Path); err != nil {
				return fmt.Errorf("could not set wallpaper: %w", err)
			}

			if err := helper.SetWallpaperMode(entry.Mode); err != nil {
				return fmt.Errorf("could not set wallpaper mode: %w", err)
			}

			if err := history.Save(histPath, h); err != nil {
				pterm.Warning.Printf("could not save history: %v\n", err)
			}

			pterm.Success.Printf("Wallpaper set to: %s\n", entry.Path)
			pterm.Info.Printf("Category: %s | Mode: %s\n", entry.Category, entry.Mode)
			return nil
		},
	}
}
