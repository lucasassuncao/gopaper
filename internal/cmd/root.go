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
	"github.com/lucasassuncao/gopaper/internal/wallhaven"
	"github.com/lucasassuncao/gopaper/internal/weather"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without subcommands
func RootCmd(g *models.Gopaper, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gopaper",
		Version: version,
		Short:   "gopaper is a CLI tool to change wallpapers based on configurable categories.",
		Long: `gopaper is a CLI tool to change wallpapers based on configurable categories.
			It allows users to define categories with specific sources and modes, and randomly selects wallpapers from enabled categories.`,
		// main.go prints the returned error itself; let that be the single
		// place the user sees it instead of also getting Cobra's own
		// "Error: ..." plus a full usage dump for every failure.
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			return preRunHandler(g, configPath)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			categoryFlag, _ := cmd.Flags().GetString("category")
			includeDisabled, _ := cmd.Flags().GetBool("include-disabled")
			return runOnce(g, categoryFlag, includeDisabled)
		},
	}
	cmd.Flags().StringP("config", "c", "", "Path to configuration file (e.g., /path/to/gopaper.yaml)")
	cmd.Flags().String("category", "", "Comma-separated category names to restrict selection to (default: all enabled categories)")
	cmd.Flags().Bool("include-disabled", false, "Include disabled categories when selecting (works with --category or alone)")
	cmd.AddCommand(EditCmd())
	cmd.AddCommand(InitCmd())
	cmd.AddCommand(PrevCmd())
	cmd.AddCommand(NextCmd())
	cmd.AddCommand(HistoryCmd())
	cmd.AddCommand(MonitorsCmd())
	cmd.AddCommand(ValidateCmd())
	cmd.AddCommand(ShowCmd())
	cmd.AddCommand(selfUpdateCmd(version))

	return cmd
}

// preRunHandler handle the pre-run configuration loading
func preRunHandler(g *models.Gopaper, configPath string) error {
	var err error
	if configPath != "" {
		dir := filepath.Dir(configPath)
		filename := filepath.Base(configPath)
		ext := filepath.Ext(filename)
		nameWithoutExt := filename[:len(filename)-len(ext)]

		err = config.InitConfig(g.Viper,
			config.WithConfigName(nameWithoutExt),
			config.WithConfigType(ext[1:]),
			config.WithConfigPath(dir),
		)
	} else {
		err = config.LoadDefault(g.Viper)
	}
	if err != nil {
		if configPath != "" {
			return fmt.Errorf("configuration file not found at '%s'", configPath)
		}
		return fmt.Errorf("configuration file not found\n\nPlease run 'gopaper init' to create a configuration file")
	}

	logger, err := config.ConfigureLogger(g.Viper)
	if err != nil {
		return fmt.Errorf("failed to configure logger: %v", err)
	}

	g.Logger = logger

	return nil
}

// runOnce picks one wallpaper from the eligible categories (or the ones
// named in categoryFlag) and applies it. g.Viper and g.Logger must already
// be initialized (via preRunHandler).
func runOnce(g *models.Gopaper, categoryFlag string, includeDisabled bool) error {
	if g.Logger == nil {
		return fmt.Errorf("logger is not initialized")
	}

	g.Logger.Info("Starting wallpaper change")

	categories, err := config.UnmarshalConfig(g)
	if err != nil {
		return err
	}
	g.Categories = categories

	candidates, err := FilterCategories(g.Categories, ParseCategoryNames(categoryFlag), includeDisabled, g.Logger)
	if err != nil {
		g.Logger.Error("invalid category selection", g.Logger.Args("error", err))
		return err
	}

	conditions, err := config.LoadConditions(g.Viper)
	if err != nil {
		g.Logger.Error("invalid conditions configuration", g.Logger.Args("error", err))
		return err
	}

	ws := fetchWeatherSnapshot(g)
	wallhavenDirs := refreshWallhavenCaches(g, candidates)
	now := time.Now()

	active := activeCategories(g, candidates, now, ws, conditions, wallhavenDirs)
	selectedCategory := helper.GetRandomCategory(active)
	if selectedCategory == nil {
		g.Logger.Error("no enabled or defined category found to select a wallpaper.")
		return fmt.Errorf("enabled categories not found")
	}

	previous, err := helper.GetPreviousWallpaper()
	if err != nil {
		g.Logger.Warn("Could not get previous wallpaper", g.Logger.Args("error", err))
	}
	selectedCategory = avoidRepeatCategory(selectedCategory, active, now, ws, conditions, wallhavenDirs, previous)

	// The drawn category decides the run's monitor mode: an "all"
	// category takes every monitor with one mirrored image (fade
	// allowed); a "per-monitor" one hands each monitor its own draw
	// among the per-monitor-eligible categories; a "monitorN" one is
	// pinned to that single monitor, leaving the others untouched.
	switch mmMode := config.MonitorModeForCategory(g.Viper, selectedCategory.MonitorOverride()); mmMode {
	case "per-monitor":
		handled, err := runPerMonitor(g, perMonitorEligible(g.Viper, active), now, ws, conditions, wallhavenDirs)
		if handled {
			return err
		}
		// Fall through to the single-wallpaper flow.
	default:
		if idx, ok := config.ParseMonitorMode(mmMode); ok {
			handled, err := runSingleMonitor(g, selectedCategory, idx, now, ws, conditions, wallhavenDirs)
			if handled {
				return err
			}
			// Fall through to the single-wallpaper flow.
		}
	}

	return applySingleWallpaper(g, selectedCategory, now, ws, conditions, wallhavenDirs, previous)
}

// activeCategories filters candidates down to the ones with a currently
// resolvable source, logging (not erroring) the ones skipped.
func activeCategories(g *models.Gopaper, candidates []*models.Categories, now time.Time, ws *weather.Snapshot, conditions map[string]models.Condition, wallhavenDirs map[*models.Categories]string) []*models.Categories {
	var active []*models.Categories
	for _, c := range candidates {
		if _, ok := helper.ResolveSource(c, now, ws, conditions, wallhavenDirs[c]); ok {
			active = append(active, c)
			continue
		}
		g.Logger.Info("Skipping category: no variant active for the current time",
			g.Logger.Args("category", c.Name))
	}
	return active
}

// avoidRepeatCategory swaps selectedCategory for a different active category
// when it would draw from the same directory as the current wallpaper and
// another active category could take its place instead.
func avoidRepeatCategory(selectedCategory *models.Categories, active []*models.Categories, now time.Time, ws *weather.Snapshot, conditions map[string]models.Condition, wallhavenDirs map[*models.Categories]string, previous string) *models.Categories {
	if len(active) <= 1 {
		return selectedCategory
	}
	resolved, ok := helper.ResolveSource(selectedCategory, now, ws, conditions, wallhavenDirs[selectedCategory])
	if !ok || config.ExpandTilde(resolved) != filepath.Dir(previous) {
		return selectedCategory
	}
	if c := helper.GetRandomCategory(excludeCategory(active, selectedCategory)); c != nil {
		return c
	}
	return selectedCategory
}

// applySingleWallpaper resolves selectedCategory's current source, picks a
// random image from it (excluding the current wallpaper when possible), and
// applies it as the single/mirrored wallpaper.
func applySingleWallpaper(g *models.Gopaper, selectedCategory *models.Categories, now time.Time, ws *weather.Snapshot, conditions map[string]models.Condition, wallhavenDirs map[*models.Categories]string, previous string) error {
	resolvedSource, _ := helper.ResolveSource(selectedCategory, now, ws, conditions, wallhavenDirs[selectedCategory])
	sourcePath := config.ExpandTilde(resolvedSource)

	files, err := helper.ReadDirectory(sourcePath)
	if err != nil {
		g.Logger.Error("error reading source directory.", g.Logger.Args("source", sourcePath, "error", err))
		return fmt.Errorf("error reading directory: %w", err)
	}

	filter, err := filters.Compile(selectedCategory.Filter)
	if err != nil {
		g.Logger.Error("invalid filter", g.Logger.Args("category", selectedCategory.Name, "error", err))
		return fmt.Errorf("invalid filter for category %q: %w", selectedCategory.Name, err)
	}

	exclude := ""
	if sourcePath == filepath.Dir(previous) {
		exclude = filepath.Base(previous)
	}
	selectedFile, err := helper.GetRandomFile(files, filter, exclude)
	if err != nil {
		g.Logger.Error("Error getting random file", g.Logger.Args("error", err))
		return fmt.Errorf("error getting random file: %w", err)
	}

	err = helper.SetWallpaperFromFile(sourcePath, selectedFile, config.TransitionEnabledForCategory(g.Viper, selectedCategory.TransitionOverride()))
	if err != nil {
		g.Logger.Error("Error setting the wallpaper", g.Logger.Args("error", err))
		return fmt.Errorf("error setting the wallpaper: %w", err)
	}

	mode := config.ModeForCategory(g.Viper, selectedCategory.ModeOverride())
	if err = helper.SetWallpaperMode(mode); err != nil {
		g.Logger.Error("Error setting wallpaper mode", g.Logger.Args("error", err))
		return fmt.Errorf("error setting wallpaper mode: %w", err)
	}

	newWallpaper := filepath.Join(sourcePath, selectedFile)

	if err := recordHistory(g, newWallpaper, selectedCategory, mode); err != nil {
		g.Logger.Warn("Could not record history", g.Logger.Args("error", err))
	}

	g.Logger.Info("Wallpaper changed successfully.",
		g.Logger.Args("category", selectedCategory.Name),
		g.Logger.Args("new wallpaper", newWallpaper),
		g.Logger.Args("previous wallpaper", previous),
		g.Logger.Args("mode", mode),
	)
	return nil
}

// excludeCategory returns cats without exclude, preserving order.
func excludeCategory(cats []*models.Categories, exclude *models.Categories) []*models.Categories {
	out := make([]*models.Categories, 0, len(cats)-1)
	for _, c := range cats {
		if c != exclude {
			out = append(out, c)
		}
	}
	return out
}

// recordHistory appends the current wallpaper to the persistent history file,
// unless configuration.history.enabled is explicitly set to false.
func recordHistory(g *models.Gopaper, wallpaper string, cat *models.Categories, mode string) error {
	return recordHistoryEntry(g, history.Entry{
		Path:      wallpaper,
		Category:  cat.Name,
		Mode:      mode,
		Timestamp: time.Now(),
	})
}

// recordHistoryEntry appends a pre-built entry (single or per-monitor) to
// the persistent history file, unless configuration.history.enabled is
// explicitly set to false.
func recordHistoryEntry(g *models.Gopaper, entry history.Entry) error {
	if !config.HistoryEnabled(g.Viper) {
		return nil
	}

	histPath, err := config.HistoryPath(g.Viper)
	if err != nil {
		return err
	}
	h, err := history.Load(histPath, config.HistoryLimit(g.Viper))
	if err != nil {
		return err
	}
	history.Append(h, entry)
	return history.Save(histPath, h)
}

// refreshWallhavenCaches resolves each wallhaven category's cache directory
// and fetches one fresh image into it. Best-effort on both counts: a
// failure only means that category runs on its existing cache (or ends up
// ineligible when the cache dir couldn't even be resolved).
func refreshWallhavenCaches(g *models.Gopaper, candidates []*models.Categories) map[*models.Categories]string {
	apiKey := config.LoadWallhavenAPIKey(g.Viper)
	dirs := map[*models.Categories]string{}
	for _, c := range candidates {
		if c.Wallhaven == nil {
			continue
		}
		dir, err := config.WallhavenCacheDir(g.Viper, c.Name, c.Wallhaven.Cache)
		if err != nil {
			g.Logger.Warn("could not resolve wallhaven cache directory, skipping category",
				g.Logger.Args("category", c.Name, "error", err))
			continue
		}
		dirs[c] = dir

		if err := wallhaven.Refresh(wallhaven.Config{
			Query:      c.Wallhaven.Query,
			Purity:     c.Wallhaven.Purity,
			APIKey:     apiKey,
			CacheLimit: c.Wallhaven.CacheLimit,
		}, dir); err != nil {
			g.Logger.Warn("wallhaven refresh failed, using cached images if any",
				g.Logger.Args("category", c.Name, "error", err))
		}
	}
	return dirs
}
