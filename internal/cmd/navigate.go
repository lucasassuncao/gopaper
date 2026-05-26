package cmd

import (
	"errors"
	"fmt"

	"github.com/lucasassuncao/gopaper/internal/helper"
	"github.com/lucasassuncao/gopaper/internal/history"
	"github.com/lucasassuncao/gopaper/internal/models"

	"github.com/pterm/pterm"
)

// navigateHistory is the shared implementation for PrevCmd and NextCmd.
// navigate is called on the loaded history to move the cursor;
// boundaryErr is the sentinel returned when the cursor is already at the edge.
func navigateHistory(navigate func(*models.History) (models.HistoryEntry, error), boundaryErr error) error {
	histPath, err := history.DefaultPath()
	if err != nil {
		return fmt.Errorf("could not determine history path: %w", err)
	}

	h, err := history.Load(histPath)
	if err != nil {
		return fmt.Errorf("could not load history: %w", err)
	}

	entry, err := navigate(h)
	if err != nil {
		if errors.Is(err, history.ErrHistoryEmpty) || errors.Is(err, boundaryErr) {
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
}
