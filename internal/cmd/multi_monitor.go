package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/lucasassuncao/gopaper/internal/config"
	"github.com/lucasassuncao/gopaper/internal/filters"
	"github.com/lucasassuncao/gopaper/internal/helper"
	"github.com/lucasassuncao/gopaper/internal/history"
	"github.com/lucasassuncao/gopaper/internal/models"
	"github.com/lucasassuncao/gopaper/internal/weather"

	"github.com/spf13/viper"
)

// runPerMonitor selects and applies a different wallpaper per monitor.
// handled reports whether the run was taken over: false means the caller
// should fall back to the single-wallpaper flow (monitor enumeration failed
// or only one monitor is connected — per-monitor is best-effort and must
// never break a working single-monitor setup).
//
// Per-monitor changes are always instant: the native crossfade relies on a
// one-item slideshow that forces the same image onto every monitor, so it
// cannot be combined with per-monitor targeting.
func runPerMonitor(g *models.Gopaper, active []*models.Categories, now time.Time, ws *weather.Snapshot, conditions map[string]models.Condition, wallhavenDirs map[*models.Categories]string) (handled bool, err error) {
	monitors, err := helper.ListMonitors()
	if err != nil {
		g.Logger.Warn("could not enumerate monitors, falling back to a single wallpaper", g.Logger.Args("error", err))
		return false, nil
	}
	if len(monitors) <= 1 {
		g.Logger.Info("only one monitor connected, using the single-wallpaper flow")
		return false, nil
	}

	var (
		targets        []helper.MonitorTarget
		monitorEntries []history.MonitorEntry
		primary        *models.Categories
		primaryPath    string
	)
	for i, devicePath := range monitors {
		candidates := categoriesForMonitor(active, i+1)
		cat := helper.GetRandomCategory(candidates)
		if cat == nil {
			g.Logger.Warn("no eligible category for monitor, leaving it unchanged", g.Logger.Args("monitor", i+1))
			continue
		}

		fullPath, err := pickWallpaperFile(cat, now, ws, conditions, wallhavenDirs[cat])
		if err != nil {
			g.Logger.Warn("could not pick a wallpaper for monitor, leaving it unchanged",
				g.Logger.Args("monitor", i+1, "category", cat.Name, "error", err))
			continue
		}

		targets = append(targets, helper.MonitorTarget{DevicePath: devicePath, Path: fullPath})
		monitorEntries = append(monitorEntries, history.MonitorEntry{Monitor: i + 1, Path: fullPath, Category: cat.Name})
		if primary == nil {
			primary = cat
			primaryPath = fullPath
		}
	}

	if len(targets) == 0 {
		g.Logger.Error("no enabled or defined category found to select a wallpaper for any monitor.")
		return true, fmt.Errorf("enabled categories not found")
	}

	if err := helper.SetWallpapersPerMonitor(targets); err != nil {
		g.Logger.Error("Error setting per-monitor wallpapers", g.Logger.Args("error", err))
		return true, fmt.Errorf("error setting the wallpaper: %w", err)
	}

	// SetPosition is global in IDesktopWallpaper — there is no per-monitor
	// position, so the primary monitor's category mode wins.
	if err := helper.SetWallpaperMode(primary.Mode); err != nil {
		g.Logger.Error("Error setting wallpaper mode", g.Logger.Args("error", err))
		return true, fmt.Errorf("error setting wallpaper mode: %w", err)
	}

	entry := history.Entry{
		Path:      primaryPath,
		Category:  primary.Name,
		Mode:      primary.Mode,
		Timestamp: time.Now(),
		Monitors:  monitorEntries,
	}
	if err := recordHistoryEntry(g, entry); err != nil {
		g.Logger.Warn("Could not record history", g.Logger.Args("error", err))
	}

	for _, m := range monitorEntries {
		g.Logger.Info("Wallpaper changed successfully.",
			g.Logger.Args("monitor", m.Monitor),
			g.Logger.Args("category", m.Category),
			g.Logger.Args("new wallpaper", m.Path),
		)
	}
	return true, nil
}

// perMonitorEligible filters active down to the categories whose effective
// multi-monitor mode is per-monitor — a "same" category only ever appears
// mirrored on every monitor, so it doesn't take part in individual draws.
func perMonitorEligible(v *viper.Viper, active []*models.Categories) []*models.Categories {
	var out []*models.Categories
	for _, c := range active {
		if config.MultiMonitorModeForCategory(v, c.MultiMonitorOverride()) == "per-monitor" {
			out = append(out, c)
		}
	}
	return out
}

// categoriesForMonitor filters active down to the categories eligible for a
// 1-based monitor index: unrestricted categories plus the ones pinned to it.
func categoriesForMonitor(active []*models.Categories, monitor int) []*models.Categories {
	var out []*models.Categories
	for _, c := range active {
		if c.Monitor == 0 || c.Monitor == monitor {
			out = append(out, c)
		}
	}
	return out
}

// pickWallpaperFile resolves a category's source directory and picks a
// random image from it, returning the image's full path.
func pickWallpaperFile(cat *models.Categories, now time.Time, ws *weather.Snapshot, conditions map[string]models.Condition, wallhavenDir string) (string, error) {
	resolvedSource, ok := helper.ResolveSource(cat, now, ws, conditions, wallhavenDir)
	if !ok {
		return "", fmt.Errorf("no active variant for category %q", cat.Name)
	}
	sourcePath := config.ExpandTilde(resolvedSource)

	files, err := helper.ReadDirectory(sourcePath)
	if err != nil {
		return "", fmt.Errorf("error reading directory: %w", err)
	}

	filter, err := filters.Compile(cat.Filter)
	if err != nil {
		return "", fmt.Errorf("invalid filter for category %q: %w", cat.Name, err)
	}

	file, err := helper.GetRandomFile(files, filter)
	if err != nil {
		return "", fmt.Errorf("error getting random file: %w", err)
	}
	return filepath.Join(sourcePath, file), nil
}
