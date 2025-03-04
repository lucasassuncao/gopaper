package models

import (
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

type Gopaper struct {
	Logger          *pterm.Logger
	Viper           *viper.Viper
	CommandFlags    *CommandFlags
	PersistentFlags *PersistentFlags
	Categories      []*Categories
}

type Categories struct {
	CategoryName string `mapstructure:"name"`
	Source       string `mapstructure:"source"`
	Mode         string `mapstructure:"mode"`
	Enabled      bool   `mapstructure:"enabled"`
}
