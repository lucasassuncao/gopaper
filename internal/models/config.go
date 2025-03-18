package models

import (
	"fmt"
	"os"
	"path/filepath"

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
	printer := pterm.DefaultInteractiveSelect.
		WithOptions([]string{"console", "log", "file", "both"}).
		WithDefaultText("Specify the output").
		WithMaxHeight(10)

	output, _ := printer.Show()

	return func(c *Config) {
		c.Configuration.Output = output
	}
}

func WithLogFile() ConfigOption {
	printer := pterm.DefaultInteractiveTextInput.
		WithDefaultText("Specify the log file").
		WithDefaultValue("C:\\logs\\gopaper.log")

	logFile, _ := printer.Show()

	return func(c *Config) {
		c.Configuration.LogFile = logFile
	}
}

func WithLogLevel() ConfigOption {
	printer := pterm.DefaultInteractiveSelect.
		WithOptions([]string{"trace", "debug", "info", "warn", "warning", "error", "fatal"}).
		WithDefaultText("Specify the log level").
		WithMaxHeight(10)

	logLevel, _ := printer.Show()

	return func(c *Config) {
		c.Configuration.LogLevel = logLevel
	}
}

func WithShowCaller() ConfigOption {
	showCaller, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("Show caller?").Show()

	return func(c *Config) {
		c.Configuration.ShowCaller = showCaller
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
		Categories: []Category{
			{
				Name:    "foo",
				Source:  "",
				Mode:    "",
				Enabled: true,
			},
			{
				Name:    "bar",
				Source:  "",
				Mode:    "",
				Enabled: false,
			},
			{
				Name:    "yin",
				Source:  "",
				Mode:    "",
				Enabled: true,
			},
			{
				Name:    "yang",
				Source:  "",
				Mode:    "",
				Enabled: false,
			},
		},
	}

	if interactive {
		applyConfigOptions(&baseConfig, configOptions)
	}

	data, err := yaml.Marshal(&baseConfig)
	if err != nil {
		return fmt.Errorf("failed to serialize yaml: %w", err)
	}

	if err := os.WriteFile(filepath.Join(path, "gopaper.yaml"), data, 0644); err != nil {
		return fmt.Errorf("failed to generate base config file: %w", err)
	}

	return nil
}

func applyConfigOptions(c *Config, configOptions []ConfigOption) {
	for _, option := range configOptions {
		option(c)
	}
}
