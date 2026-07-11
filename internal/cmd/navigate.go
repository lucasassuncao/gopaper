package cmd

import (
	"errors"
	"fmt"

	"github.com/lucasassuncao/gopaper/internal/config"
	"github.com/lucasassuncao/gopaper/internal/helper"
	"github.com/lucasassuncao/gopaper/internal/history"
	"github.com/lucasassuncao/gopaper/internal/models"

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

	if err := applyHistoryEntry(v, entry); err != nil {
		return err
	}

	if err := history.Save(histPath, h); err != nil {
		logger.Warn("could not save history", logger.Args("error", err))
	}

	logger.Info("Wallpaper changed successfully.",
		logger.Args("wallpaper", entry.Path, "category", entry.Category, "mode", entry.Mode),
	)
	return nil
}

// applyHistoryEntry re-applies a history entry to the desktop: per-monitor
// when the entry recorded monitors, otherwise the regular single-wallpaper
// path honoring the entry's category transition override (looked up by name
// in the current config — the category may have been edited or removed
// since the entry was recorded, in which case the global setting applies).
func applyHistoryEntry(v *viper.Viper, entry history.Entry) error {
	if err := applyEntryWallpaper(v, entry); err != nil {
		return fmt.Errorf("could not set wallpaper: %w", err)
	}
	if err := helper.SetWallpaperMode(entry.Mode); err != nil {
		return fmt.Errorf("could not set wallpaper mode: %w", err)
	}
	return nil
}

// applyEntryWallpaper puts the entry's image(s) on the desktop: per-monitor
// when recorded that way, otherwise the single-wallpaper path.
func applyEntryWallpaper(v *viper.Viper, entry history.Entry) error {
	if len(entry.Monitors) > 0 {
		return applyMonitorsEntry(entry)
	}
	return helper.SetWallpaperFromPath(entry.Path, config.TransitionEnabledForCategory(v, categoryTransition(v, entry.Category)))
}

// categoryTransition returns the transition override of the named category
// in the current config, or "" when the category no longer exists (the
// global setting then applies).
func categoryTransition(v *viper.Viper, name string) string {
	var categories []*models.Categories
	if err := v.UnmarshalKey("categories", &categories); err != nil {
		return ""
	}
	for _, c := range categories {
		if c.Name == name {
			return c.TransitionOverride()
		}
	}
	return ""
}

// applyMonitorsEntry re-applies a per-monitor history entry by matching each
// recorded 1-based monitor index against the monitors present now (device
// paths are not persisted). Entries whose monitor is gone are skipped with a
// warning; it errors only when none can be applied.
func applyMonitorsEntry(entry history.Entry) error {
	monitors, err := helper.ListMonitors()
	if err != nil {
		return err
	}

	var targets []helper.MonitorTarget
	for _, m := range entry.Monitors {
		idx := m.Monitor - 1
		if idx < 0 || idx >= len(monitors) {
			logger.Warn("skipping monitor from history entry: not connected now", logger.Args("monitor", m.Monitor))
			continue
		}
		targets = append(targets, helper.MonitorTarget{DevicePath: monitors[idx], Path: m.Path})
	}
	if len(targets) == 0 {
		return fmt.Errorf("none of the entry's monitors are connected")
	}
	return helper.SetWallpapersPerMonitor(targets)
}
