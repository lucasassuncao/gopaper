package models

import (
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

type Config struct {
	Configuration Configuration `yaml:"configuration"`
	Categories    []Categories  `yaml:"categories"`
}

type Configuration struct {
	Output     string `yaml:"output"`
	LogFile    string `yaml:"log-file"`
	LogLevel   string `yaml:"log-level"`
	ShowCaller bool   `yaml:"show-caller"`
}

type Gopaper struct {
	Logger     *pterm.Logger
	Viper      *viper.Viper
	Categories []*Categories
}

type Categories struct {
	Name    string `yaml:"name" mapstructure:"name"`
	Source  string `yaml:"source" mapstructure:"source"`
	Mode    string `yaml:"mode" mapstructure:"mode"`
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
}
