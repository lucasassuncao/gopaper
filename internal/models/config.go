package models

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/pterm/pterm"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Configuration Configuration `yaml:"configuration"`
	Categories    []Category    `yaml:"categories"`
}

type Configuration struct {
	Output     string `yaml:"output"`
	LogFile    string `yaml:"log-file"`
	LogLevel   string `yaml:"log-level"`
	ShowCaller bool   `yaml:"show-caller"`
}
type Category struct {
	Name    string `yaml:"name"`
	Source  string `yaml:"source"`
	Mode    string `yaml:"mode"`
	Enabled bool   `yaml:"enabled"`
}

type ConfigOption func(*Config)

func WithOutput() ConfigOption {
	clearScreen()
	output, _ := pterm.DefaultInteractiveSelect.WithOptions([]string{"console", "log", "file", "both"}).WithDefaultText("Specify the output").WithMaxHeight(10).Show()

	return func(c *Config) {
		c.Configuration.Output = output
	}
}

func WithLogFile() ConfigOption {
	clearScreen()
	logFile, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Specify the log file").WithDefaultValue("C:\\logs\\gopaper.log").Show()

	return func(c *Config) {
		c.Configuration.LogFile = logFile
	}
}

func WithLogLevel() ConfigOption {
	clearScreen()
	logLevel, _ := pterm.DefaultInteractiveSelect.WithOptions([]string{"trace", "debug", "info", "warn", "warning", "error", "fatal"}).WithDefaultText("Specify the log level").WithMaxHeight(10).Show()

	return func(c *Config) {
		c.Configuration.LogLevel = logLevel
	}
}

func WithShowCaller() ConfigOption {
	clearScreen()
	showCaller, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("Show caller?").Show()

	return func(c *Config) {
		c.Configuration.ShowCaller = showCaller
	}
}

func WithCategory() ConfigOption {
	clearScreen()
	var categories []Category

	want, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("Do you want to add categories?").Show()

	if want {
		for {
			clearScreen()
			name, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Specify the category name").Show()
			source, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Specify the source directory").Show()
			mode, _ := pterm.DefaultInteractiveSelect.WithOptions([]string{"crop", "tile", "stretch", "span", "fit", "center"}).WithDefaultText("Specify the wallpaper mode").WithMaxHeight(10).Show()
			enabled, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("Enable wallpaper category?").Show()

			categories = append(categories, Category{
				Name:    name,
				Source:  source,
				Mode:    mode,
				Enabled: enabled,
			})

			addMore, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("Do you want to add another category?").Show()
			if !addMore {
				break
			}
		}
	}

	return func(c *Config) {
		c.Categories = append(c.Categories, categories...)
	}
}

func NewConfig(path string, interactive bool, configOptions ...ConfigOption) error {
	baseConfig := Config{
		Configuration: Configuration{
			Output:     "",
			LogFile:    "",
			LogLevel:   "",
			ShowCaller: false,
		},
		Categories: []Category{},
	}

	if interactive {
		applyConfigOptions(&baseConfig, configOptions)
	}

	if len(baseConfig.Categories) == 0 {
		baseConfig.Categories = append(baseConfig.Categories, Category{
			Name:    "default",
			Source:  "C:\\wallpapers",
			Mode:    "crop",
			Enabled: true,
		})
	}

	data, err := yaml.Marshal(&baseConfig)
	if err != nil {
		return fmt.Errorf("failed to serialize yaml: %w", err)
	}

	if err := os.WriteFile(filepath.Join(path, "gopaper.yaml"), data, 0644); err != nil {
		return fmt.Errorf("failed to generate base config file: %w", err)
	}

	clearScreen()
	return nil
}

func applyConfigOptions(c *Config, configOptions []ConfigOption) {
	for _, option := range configOptions {
		option(c)
	}
}

// clearScreen clears the terminal screen based on the operating system
func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}
