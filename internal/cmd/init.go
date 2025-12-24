package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucasassuncao/gopaper/internal/helper"
	"github.com/lucasassuncao/gopaper/internal/models"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var (
	initForce       bool
	initInteractive bool
	initTemplate    string
)

// InitCmd generates a configuration file
func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize gopaper configuration",
		Long: `Initialize gopaper configuration file with predefined templates or interactive mode.

Available templates:
  - full:     Complete example with multiple categories

The configuration file will be created at: <executable_dir>/conf/gopaper.yaml`,
		Example: `  # Interactive mode (recommended for first time)
  gopaper init -i

  # Use a template
  gopaper init -t full

  # Force overwrite existing config
  gopaper init -f`,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ex, err := os.Executable()
		if err != nil {
			return fmt.Errorf("error getting executable: %v", err)
		}

		configPath := filepath.Join(filepath.Dir(ex), "conf")
		configFile := filepath.Join(configPath, "gopaper.yaml")

		if _, err := os.Stat(configFile); err == nil && !initForce {
			pterm.Error.Printf("Configuration file already exists at: %s\n", configFile)
			pterm.Info.Println("Use --force to overwrite")
			return nil
		}

		if err := helper.CreateDirectory(configPath); err != nil {
			return fmt.Errorf("error creating config directory: %v", err)
		}

		var config *models.Config

		if initInteractive {
			config = generateInteractiveConfig()
		} else {
			config = getTemplateConfig(initTemplate)
		}

		// Write config to file
		data, err := yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("error marshaling config: %v", err)
		}

		if err := os.WriteFile(configFile, data, 0644); err != nil {
			return fmt.Errorf("error writing config file: %v", err)
		}

		clearScreen()
		pterm.Success.Printf("Configuration file created at: %s\n", configFile)
		pterm.Info.Println("\nNext steps:")
		pterm.Info.Println("  1. Edit the configuration file to customize categories")
		pterm.Info.Println("  2. Add your wallpaper images to the source directories")
		pterm.Info.Println("  3. Run 'gopaper' to change your wallpaper")

		return nil
	}

	cmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing configuration file")
	cmd.Flags().BoolVarP(&initInteractive, "interactive", "i", false, "Interactive mode with prompts")
	cmd.Flags().StringVarP(&initTemplate, "template", "t", "basic", "Template to use (basic, nature, mixed, full)")

	return cmd
}

// generateInteractiveConfig creates configuration through interactive prompts
func generateInteractiveConfig() *models.Config {
	clearScreen()
	pterm.DefaultHeader.WithFullWidth().Println("Gopaper Configuration Generator")
	pterm.Println()

	config := &models.Config{
		Configuration: models.Configuration{},
		Categories:    []models.Categories{},
	}

	pterm.DefaultSection.Println("Logging Configuration")
	output, _ := pterm.DefaultInteractiveSelect.
		WithOptions([]string{"console", "log", "file", "both", "none"}).
		WithDefaultText("Where should logs be output?").
		WithMaxHeight(10).
		Show()
	config.Configuration.Output = output

	if output == "log" || output == "file" || output == "both" {
		defaultLogPath := getDefaultLogPath()
		logFile, _ := pterm.DefaultInteractiveTextInput.
			WithDefaultText("Log file path").
			WithDefaultValue(defaultLogPath).
			Show()
		config.Configuration.LogFile = logFile
	}

	logLevel, _ := pterm.DefaultInteractiveSelect.
		WithOptions([]string{"trace", "debug", "info", "warn", "error", "fatal"}).
		WithDefaultText("Log level").
		WithDefaultOption("info").
		WithMaxHeight(10).
		Show()
	config.Configuration.LogLevel = logLevel

	showCaller, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultText("Show caller information in logs?").
		WithDefaultValue(false).
		Show()
	config.Configuration.ShowCaller = showCaller

	clearScreen()
	pterm.DefaultSection.Println("Categories Configuration")
	pterm.Info.Println("Categories define wallpaper collections")
	pterm.Println()

	addCategories, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultText("Do you want to add categories now?").
		WithDefaultValue(true).
		Show()

	if addCategories {
		config.Categories = collectCategories()
	}

	// Add default category if none were added
	if len(config.Categories) == 0 {
		config.Categories = append(config.Categories, getDefaultCategory())
	}

	return config
}

// collectCategories collects categories from user input
func collectCategories() []models.Categories {
	var categories []models.Categories

	for {
		clearScreen()
		pterm.DefaultHeader.WithFullWidth().Printf("Category #%d", len(categories)+1)
		pterm.Println()

		name, _ := pterm.DefaultInteractiveTextInput.
			WithDefaultText("Category name (e.g., Custom Selection, Wallhaven, Girls)").
			Show()

		if name == "" {
			pterm.Warning.Println("Category name is required")
			continue
		}

		enabled, _ := pterm.DefaultInteractiveConfirm.
			WithDefaultText("Enable this category?").
			WithDefaultValue(true).
			Show()

		source, _ := pterm.DefaultInteractiveTextInput.
			WithDefaultText("Source directory (where wallpapers are stored)").
			WithDefaultValue(getDefaultSourcePath(name)).
			Show()

		mode, _ := pterm.DefaultInteractiveSelect.
			WithOptions([]string{"crop", "tile", "stretch", "span", "fit", "center"}).
			WithDefaultText("Wallpaper display mode").
			WithDefaultOption("crop").
			Show()

		category := models.Categories{
			Name:    name,
			Enabled: enabled,
			Source:  source,
			Mode:    mode,
		}

		categories = append(categories, category)

		// Summary
		pterm.Println()
		pterm.DefaultSection.Println("Category Summary")
		printCategorySummary(category)
		pterm.Println()

		// Add more?
		addMore, _ := pterm.DefaultInteractiveConfirm.
			WithDefaultText("Add another category?").
			WithDefaultValue(false).
			Show()

		if !addMore {
			break
		}
	}
	return categories
}

// getDefaultCategory returns a default category configuration
func getDefaultCategory() models.Categories {
	return models.Categories{
		Name:    "Custom Selection",
		Enabled: true,
		Source:  getDefaultSourcePath("CustomSelection"),
		Mode:    "crop",
	}
}

// getTemplateConfig returns a predefined template configuration
func getTemplateConfig(template string) *models.Config {
	templates := map[string]func() *models.Config{
		"full": getFullTemplate,
	}

	templateFunc, exists := templates[template]
	if !exists {
		pterm.Warning.Printf("Unknown template '%s', using 'basic'\n", template)
		templateFunc = getBasicTemplate
	}

	return templateFunc()
}

// getBasicTemplate returns the basic configuration template
func getBasicTemplate() *models.Config {
	return &models.Config{
		Configuration: models.Configuration{
			Output:     "console",
			LogLevel:   "info",
			ShowCaller: false,
		},
		Categories: []models.Categories{
			{
				Name:    "Custom Selection",
				Source:  getDefaultSourcePath("CustomSelection"),
				Mode:    "crop",
				Enabled: true,
			},
		},
	}
}

// getFullTemplate returns the full configuration template
func getFullTemplate() *models.Config {
	return &models.Config{
		Configuration: models.Configuration{
			Output:     "both",
			LogFile:    getDefaultLogPath(),
			LogLevel:   "info",
			ShowCaller: false,
		},
		Categories: []models.Categories{
			{
				Name:    "Custom Selection",
				Source:  getDefaultSourcePath("CustomSelection"),
				Mode:    "crop",
				Enabled: true,
			},
			{
				Name:    "Wallhaven",
				Source:  getDefaultSourcePath("Wallhaven"),
				Mode:    "crop",
				Enabled: false,
			},
			{
				Name:    "Wallbase",
				Source:  getDefaultSourcePath("Wallbase"),
				Mode:    "crop",
				Enabled: false,
			},
			{
				Name:    "Wallpapers Wide",
				Source:  getDefaultSourcePath("WallpapersWide"),
				Mode:    "crop",
				Enabled: false,
			},
			{
				Name:    "UHD Wallpaper",
				Source:  getDefaultSourcePath("UHDPaper"),
				Mode:    "crop",
				Enabled: false,
			},
			{
				Name:    "Looking at Viewer",
				Source:  filepath.Join(getDefaultSourcePath("Wallhaven"), "Girls", "Looking At Viewer"),
				Mode:    "crop",
				Enabled: false,
			},
			{
				Name:    "Girls",
				Source:  filepath.Join(getDefaultSourcePath("Wallhaven"), "Girls"),
				Mode:    "crop",
				Enabled: false,
			},
			{
				Name:    "Better Sonoma",
				Source:  filepath.Join(getDefaultSourcePath("OS"), "Apple", "BasicAppleGuy", "Better Sonoma"),
				Mode:    "crop",
				Enabled: false,
			},
			{
				Name:    "Saltern Study",
				Source:  filepath.Join(getDefaultSourcePath("OS"), "Apple", "BasicAppleGuy", "Saltern Study"),
				Mode:    "crop",
				Enabled: false,
			},
			{
				Name:    "Saltern Study Night",
				Source:  filepath.Join(getDefaultSourcePath("OS"), "Apple", "BasicAppleGuy", "Saltern Study Night"),
				Mode:    "crop",
				Enabled: false,
			},
		},
	}
}

// getDefaultSourcePath returns the default source path for wallpapers
func getDefaultSourcePath(categoryName string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("C:\\", "Users", "Public", "Pictures", "Walls", categoryName)
	}
	return filepath.Join(homeDir, "Imagens", "Walls", categoryName)
}

// getDefaultLogPath returns the default log file path
func getDefaultLogPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "C:\\logs\\gopaper.log"
	}
	return filepath.Join(homeDir, "Documents", "Scripts", "go", "logs", "gopaper.log")
}

// printCategorySummary prints a summary of the category configuration
func printCategorySummary(category models.Categories) {
	enabledStr := "No"
	if category.Enabled {
		enabledStr = "Yes"
	}

	pterm.Printf("  Name:    %s\n", pterm.Cyan(category.Name))
	pterm.Printf("  Enabled: %s\n", pterm.Magenta(enabledStr))
	pterm.Printf("  Mode:    %s\n", pterm.Yellow(category.Mode))
	pterm.Printf("  Source:  %s\n", pterm.Yellow(category.Source))
}

// clearScreen clears the terminal screen
func clearScreen() {
	pterm.Println()
}
