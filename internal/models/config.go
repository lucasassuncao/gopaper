package models

import (
	"fmt"
	"os"
	"path/filepath"

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

func NewConfig(path string) error {
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

	data, err := yaml.Marshal(&baseConfig)
	if err != nil {
		return fmt.Errorf("failed to serialize yaml: %w", err)
	}

	if err := os.WriteFile(filepath.Join(path, "movelooper.yaml"), data, 0644); err != nil {
		return fmt.Errorf("failed to generate base config file: %w", err)
	}

	return nil
}
