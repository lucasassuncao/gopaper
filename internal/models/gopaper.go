package models

import (
	"time"

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

// HistoryEntry represents a single wallpaper that was applied.
type HistoryEntry struct {
	Path      string    `json:"path"`
	Category  string    `json:"category"`
	Mode      string    `json:"mode"`
	Timestamp time.Time `json:"timestamp"`
}

// History is the persistent navigation state for wallpaper history.
// Entries are stored newest-first (index 0 = most recently applied).
// CurrentIndex tracks which entry is currently displayed.
type History struct {
	Entries      []HistoryEntry `json:"entries"`
	CurrentIndex int            `json:"current_index"`
	MaxEntries   int            `json:"max_entries"`
}
