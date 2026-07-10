package cmd

import (
	"errors"
	"fmt"

	"github.com/lucasassuncao/gopaper/internal/config"
	"github.com/lucasassuncao/gopaper/internal/helper"
	"github.com/lucasassuncao/gopaper/internal/history"

	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

var logger = pterm.DefaultLogger.WithLevel(pterm.LogLevelTrace)

// navigateHistory is the shared implementation for PrevCmd and NextCmd.
// navigate is called on the loaded history to move the cursor;
// boundaryErr is the sentinel returned when the cursor is already at the edge.
func navigateHistory(navigate func(*history.History) (history.Entry, error), boundaryErr error) error {
	v := viper.GetViper()
	if err := config.LoadDefault(v); err != nil {
		var notFound config.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("could not load configuration: %w", err)
		}
		// No config file found: fall back to the built-in history defaults.
	}

	histPath, err := config.HistoryPath(v)
	if err != nil {
		return fmt.Errorf("could not determine history path: %w", err)
	}

	h, err := history.Load(histPath, config.HistoryLimit(v))
	if err != nil {
		return fmt.Errorf("could not load history: %w", err)
	}

	entry, err := navigate(h)
	if err != nil {
		if errors.Is(err, history.ErrHistoryEmpty) || errors.Is(err, boundaryErr) {
			logger.Warn(err.Error())
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
		logger.Warn("could not save history", logger.Args("error", err))
	}

	logger.Info("Wallpaper changed successfully.",
		logger.Args("wallpaper", entry.Path, "category", entry.Category, "mode", entry.Mode),
	)
	return nil
}
