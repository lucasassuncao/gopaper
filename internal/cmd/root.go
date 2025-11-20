package cmd

import (
	"fmt"
	"gopaper/internal/config"
	"gopaper/internal/helper"
	"gopaper/internal/models"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without subcommands
func RootCmd(g *models.Gopaper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gopaper",
		Short: "gopaper is a CLI tool to change wallpapers based on configurable categories.",
		Long: `gopaper is a CLI tool to change wallpapers based on configurable categories.
			It allows users to define categories with specific sources and modes, and randomly selects wallpapers from enabled categories.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			return preRunHandler(g, configPath)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if g.Logger == nil {
				return fmt.Errorf("logger is not initialized")
			}

			g.Logger.Info("Starting wallpaper change")

			g.Categories = config.UnmarshalConfig(g)

			enabledCategories := helper.GetEnabledCategories(g.Categories)

			selectedCategory := helper.GetRandomCategory(enabledCategories)
			if selectedCategory == nil {
				g.Logger.Error("no enabled or defined category found to select a wallpaper.")
				return fmt.Errorf("enabled categories not found")
			}

			files, err := helper.ReadDirectory(selectedCategory.Source)
			if err != nil {
				g.Logger.Error("error reading source directory.", g.Logger.Args("source", selectedCategory.Source, "error", err))
				return fmt.Errorf("error reading directory.: %w", err)
			}

			selectedFile, err := helper.GetRandomFile(files)
			if err != nil {
				g.Logger.Error("Error getting random file", g.Logger.Args("error", err))
				return fmt.Errorf("error getting random file: %w", err)
			}

			previous, _ := helper.GetPreviousWallpaper()

			err = helper.SetWallpaperFromFile(selectedCategory.Source, selectedFile)
			if err != nil {
				g.Logger.Error("Error setting the wallpaper", g.Logger.Args("error", err))
				return fmt.Errorf("error setting the wallpaper: %w", err)
			}

			helper.SetWallpaperMode(selectedCategory.Mode)

			g.Logger.Info("Wallpaper changed successfully.",
				g.Logger.Args("category", selectedCategory.Name),
				g.Logger.Args("new wallpaper", filepath.Join(selectedCategory.Source, selectedFile)),
				g.Logger.Args("previous wallpaper", previous),
				g.Logger.Args("mode", selectedCategory.Mode),
			)
			return nil
		},
	}
	cmd.Flags().StringP("config", "c", "", "Path to configuration file (e.g., /path/to/gopaper.yaml)")
	cmd.AddCommand(InitCmd())

	return cmd
}

// preRunHandler handle the pre-run configuration loading
func preRunHandler(g *models.Gopaper, configPath string) error {
	var options []config.ViperOptions

	if configPath != "" {
		// Se um caminho específico foi fornecido, use-o
		dir := filepath.Dir(configPath)
		filename := filepath.Base(configPath)
		ext := filepath.Ext(filename)
		nameWithoutExt := filename[:len(filename)-len(ext)]

		options = []config.ViperOptions{
			config.WithConfigName(nameWithoutExt),
			config.WithConfigType(ext[1:]),
			config.WithConfigPath(dir),
		}
	} else {
		ex, err := os.Executable()
		if err != nil {
			return fmt.Errorf("error getting executable: %v", err)
		}

		options = []config.ViperOptions{
			config.WithConfigName("gopaper"),
			config.WithConfigType("yaml"),
			config.WithConfigPath(filepath.Dir(ex)),
			config.WithConfigPath(filepath.Join(filepath.Dir(ex), "conf")),
		}
	}

	err := config.InitConfig(g.Viper, options...)
	if err != nil {
		if configPath != "" {
			return fmt.Errorf("configuration file not found at '%s'", configPath)
		}
		return fmt.Errorf("configuration file not found\n\nPlease run 'movelooper init' to create a configuration file")
	}

	logger, err := config.ConfigureLogger(g.Viper)
	if err != nil {
		return fmt.Errorf("failed to configure logger: %v", err)
	}

	g.Logger = logger

	return nil
}
