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
			if g.Logger == nil {
				return fmt.Errorf("logger is not initialized")
			}

			g.Logger.Info("Starting wallpaper change")

			categories, err := config.UnmarshalConfig(g)
			if err != nil {
				return err
			}
			g.Categories = categories

			categoryFlag, _ := cmd.Flags().GetString("category")
			includeDisabled, _ := cmd.Flags().GetBool("include-disabled")

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

			now := time.Now()
			var active []*models.Categories
			for _, c := range candidates {
				if _, ok := helper.ResolveSource(c, now, ws, conditions); ok {
					active = append(active, c)
					continue
				}
				g.Logger.Info("Skipping category: no variant active for the current time",
					g.Logger.Args("category", c.Name))
			}

			selectedCategory := helper.GetRandomCategory(active)
			if selectedCategory == nil {
				g.Logger.Error("no enabled or defined category found to select a wallpaper.")
				return fmt.Errorf("enabled categories not found")
			}

			resolvedSource, _ := helper.ResolveSource(selectedCategory, now, ws, conditions)
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

			selectedFile, err := helper.GetRandomFile(files, filter)
			if err != nil {
				g.Logger.Error("Error getting random file", g.Logger.Args("error", err))
				return fmt.Errorf("error getting random file: %w", err)
			}

			previous, err := helper.GetPreviousWallpaper()
			if err != nil {
				g.Logger.Warn("Could not get previous wallpaper", g.Logger.Args("error", err))
			}

			err = helper.SetWallpaperFromFile(sourcePath, selectedFile, config.TransitionEnabled(g.Viper))
			if err != nil {
				g.Logger.Error("Error setting the wallpaper", g.Logger.Args("error", err))
				return fmt.Errorf("error setting the wallpaper: %w", err)
			}

			if err = helper.SetWallpaperMode(selectedCategory.Mode); err != nil {
				g.Logger.Error("Error setting wallpaper mode", g.Logger.Args("error", err))
				return fmt.Errorf("error setting wallpaper mode: %w", err)
			}

			newWallpaper := filepath.Join(sourcePath, selectedFile)

			if err := recordHistory(g, newWallpaper, selectedCategory); err != nil {
				g.Logger.Warn("Could not record history", g.Logger.Args("error", err))
			}

			g.Logger.Info("Wallpaper changed successfully.",
				g.Logger.Args("category", selectedCategory.Name),
				g.Logger.Args("new wallpaper", newWallpaper),
				g.Logger.Args("previous wallpaper", previous),
				g.Logger.Args("mode", selectedCategory.Mode),
			)
			return nil
		},
	}
	cmd.Flags().StringP("config", "c", "", "Path to configuration file (e.g., /path/to/gopaper.yaml)")
	cmd.Flags().String("category", "", "Comma-separated category names to restrict selection to (default: all enabled categories)")
	cmd.Flags().Bool("include-disabled", false, "Include disabled categories when selecting (works with --category or alone)")
	cmd.AddCommand(EditCmd())
	cmd.AddCommand(InitCmd())
	cmd.AddCommand(PrevCmd())
	cmd.AddCommand(NextCmd())
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

// recordHistory appends the current wallpaper to the persistent history file,
// unless configuration.history.enabled is explicitly set to false.
func recordHistory(g *models.Gopaper, wallpaper string, cat *models.Categories) error {
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
	history.Append(h, history.Entry{
		Path:      wallpaper,
		Category:  cat.Name,
		Mode:      cat.Mode,
		Timestamp: time.Now(),
	})
	return history.Save(histPath, h)
}
