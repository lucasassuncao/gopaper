package cmd

import (
	"fmt"
	"gopaper/internal/config"
	"gopaper/internal/helper"
	"gopaper/internal/models"
	"path/filepath"

	"github.com/spf13/cobra"
)

// ChangeCmd represents the change command
func ChangeCmd(g *models.Gopaper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change",
		Short: "Change the wallpaper of the Windows Desktop",
		Long:  "Changes the wallpaper of the Windows Desktop based on the specified category",
	}

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if g.Logger == nil {
			return fmt.Errorf("logger is not initialized")
		}

		g.Logger.Info("Changing Wallpaper")
		g.Logger.Debug("Using Configuration",
			g.Logger.Args("output", *g.PersistentFlags.Output),
			g.Logger.Args("show-caller", *g.PersistentFlags.ShowCaller),
			g.Logger.Args("log-level", *g.PersistentFlags.LogLevel),
			g.Logger.Args("log-file", g.Viper.GetString("configuration.log-file")),
			g.Logger.Args("config-file", g.Viper.ConfigFileUsed()),
		)

		return nil
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		g.Categories = config.UnmarshalConfig(g)

		enabledCategories := helper.GetEnabledCategories(g.Categories)
		selectedCategory := helper.GetRandomCategory(enabledCategories)

		files, err := helper.ReadDirectory(selectedCategory.Source)
		if err != nil {
			return fmt.Errorf("error reading directory: %v", err)
		}

		selectedFile, err := helper.GetRandomFile(files)
		if err != nil {
			return fmt.Errorf("error getting random file: %v", err)
		}

		previous, err := helper.GetPreviousWallpaper()
		if err != nil {
			return fmt.Errorf("error getting previous wallpaper: %v", err)
		}

		err = helper.SetWallpaperFromFile(selectedCategory.Source, selectedFile)
		if err != nil {
			return fmt.Errorf("error setting wallpaper: %v", err)
		}

		helper.SetWallpaperMode(selectedCategory.Mode)

		g.Logger.Info("Wallpaper Changed Successfully",
			g.Logger.Args("category", selectedCategory.CategoryName),
			g.Logger.Args("mode", selectedCategory.Mode),
			g.Logger.Args("previous wallpaper", previous),
			g.Logger.Args("new wallpaper", filepath.Join(selectedCategory.Source, selectedFile)),
		)

		return nil
	}
	return cmd
}
