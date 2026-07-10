package models

import (
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/viper"

	"github.com/lucasassuncao/yedit/editor"
	"github.com/lucasassuncao/yedit/metadata"
)

type Config struct {
	Configuration Configuration `yaml:"configuration" mapstructure:"configuration"`
	Categories    []Categories  `yaml:"categories" mapstructure:"categories"`
}

// Configuration holds the general settings for gopaper, grouped into logging
// and history sub-sections.
type Configuration struct {
	Logging Logging `yaml:"logging" mapstructure:"logging"`
	History History `yaml:"history" mapstructure:"history"`
}

// Logging holds the log output settings.
type Logging struct {
	Output     string `yaml:"output" mapstructure:"output"`
	Level      string `yaml:"level" mapstructure:"level"`
	File       string `yaml:"file" mapstructure:"file"`
	ShowCaller bool   `yaml:"show-caller" mapstructure:"show-caller"`
}

// History holds the wallpaper history settings (used by the prev/next
// commands).
type History struct {
	Limit   int    `yaml:"limit,omitempty" mapstructure:"limit"`
	File    string `yaml:"file,omitempty" mapstructure:"file"`
	Enabled bool   `yaml:"enabled,omitempty" mapstructure:"enabled"`
}

func (Config) Metadata() map[string]*metadata.Node {
	return map[string]*metadata.Node{
		"configuration": {FieldMeta: editor.FieldMeta{
			Description: "General settings for gopaper, grouped into logging and history sub-sections.",
			Required:    true,
		}},
		"categories": {FieldMeta: editor.FieldMeta{
			Description: "List of wallpaper categories. Each entry defines a source directory, a display mode, and whether it is enabled.",
			Required:    true,
			MinCount:    1,
		}},
	}
}

func (Configuration) Metadata() map[string]*metadata.Node {
	return map[string]*metadata.Node{
		"logging": {FieldMeta: editor.FieldMeta{
			Description: "Log output settings: destination, severity level, file path, and caller info.",
			Required:    true,
		}},
		"history": {FieldMeta: editor.FieldMeta{
			Description: "Wallpaper history settings, used by the prev/next commands.",
		}},
	}
}

func (Logging) Metadata() map[string]*metadata.Node {
	return map[string]*metadata.Node{
		"output": {FieldMeta: editor.FieldMeta{
			Description: "Where log output is written.",
			Required:    true,
			OneOf:       []string{"console", "log", "file", "both", "none"},
			Default:     "console",
		}},
		"level": {FieldMeta: editor.FieldMeta{
			Description: "Minimum severity level logged.",
			Required:    true,
			OneOf:       []string{"trace", "debug", "info", "warn", "error", "fatal"},
			Default:     "info",
		}},
		"file": {FieldMeta: editor.FieldMeta{
			Description: "Path to the log file. Required when output is \"log\", \"file\", or \"both\".",
			Example:     `file: "C:\logs\gopaper.log"`,
		}},
		"show-caller": {FieldMeta: editor.FieldMeta{
			Description: "Include the calling file and line number in log output.",
			Default:     "false",
		}},
	}
}

func (History) Metadata() map[string]*metadata.Node {
	return map[string]*metadata.Node{
		"limit": {FieldMeta: editor.FieldMeta{
			Description: "Maximum number of wallpapers kept in history for prev/next.",
			Min:         "1",
			Default:     "50",
		}},
		"file": {FieldMeta: editor.FieldMeta{
			Description: "Path to the history file. Defaults to a history/ subdirectory next to the executable.",
		}},
		"enabled": {FieldMeta: editor.FieldMeta{
			Description: "Whether wallpaper changes are recorded to history.",
			Default:     "true",
		}},
	}
}

func (Categories) Metadata() map[string]*metadata.Node {
	return map[string]*metadata.Node{
		"name": {FieldMeta: editor.FieldMeta{
			Description: "Unique display name for this category.",
			Required:    true,
			Unique:      true,
		}},
		"source": {FieldMeta: editor.FieldMeta{
			Description: "Directory containing the wallpaper images for this category.",
			Required:    true,
		}},
		"mode": {FieldMeta: editor.FieldMeta{
			Description: "Wallpaper display mode applied when an image from this category is selected.",
			Required:    true,
			OneOf:       []string{"crop", "tile", "stretch", "span", "fit", "center"},
			Default:     "crop",
		}},
		"enabled": {FieldMeta: editor.FieldMeta{
			Description: "Whether this category is eligible for random selection.",
			Default:     "true",
		}},
		"filter": {FieldMeta: editor.FieldMeta{
			Description: "Optional constraints narrowing which files in source are eligible, beyond the fixed image-extension check.",
		}},
	}
}

type Gopaper struct {
	Logger     *pterm.Logger
	Viper      *viper.Viper
	Categories []*Categories
}

type Categories struct {
	Name    string  `yaml:"name" mapstructure:"name"`
	Source  string  `yaml:"source" mapstructure:"source"`
	Mode    string  `yaml:"mode" mapstructure:"mode"`
	Enabled bool    `yaml:"enabled" mapstructure:"enabled"`
	Filter  *Filter `yaml:"filter,omitempty" mapstructure:"filter"`
}

// Filter narrows which files in a category's source directory are eligible
// for selection, beyond the fixed image-extension check. Match, Age, and Size
// combine with AND semantics; a nil sub-filter imposes no constraint.
type Filter struct {
	Match *MatchFilter `yaml:"match,omitempty" mapstructure:"match"`
	Age   *AgeFilter   `yaml:"age,omitempty" mapstructure:"age"`
	Size  *SizeFilter  `yaml:"size,omitempty" mapstructure:"size"`
}

// MatchFilter matches a file by its name. Literal, Regex, and Glob are
// mutually exclusive.
type MatchFilter struct {
	Literal       string `yaml:"literal,omitempty" mapstructure:"literal"`
	Regex         string `yaml:"regex,omitempty" mapstructure:"regex"`
	Glob          string `yaml:"glob,omitempty" mapstructure:"glob"`
	CaseSensitive bool   `yaml:"case-sensitive,omitempty" mapstructure:"case-sensitive"`
}

// AgeFilter matches a file by how long ago it was last modified.
type AgeFilter struct {
	Min time.Duration `yaml:"min,omitempty" mapstructure:"min"`
	Max time.Duration `yaml:"max,omitempty" mapstructure:"max"`
}

// SizeFilter matches a file by its size in bytes. Min/Max accept
// human-readable strings such as "10MB" or "1.5GiB".
type SizeFilter struct {
	Min string `yaml:"min,omitempty" mapstructure:"min"`
	Max string `yaml:"max,omitempty" mapstructure:"max"`
}

func (Filter) Metadata() map[string]*metadata.Node {
	return map[string]*metadata.Node{
		"match": {FieldMeta: editor.FieldMeta{
			Description: "Match files by name.",
		}},
		"age": {FieldMeta: editor.FieldMeta{
			Description: "Match files by how long ago they were last modified.",
		}},
		"size": {FieldMeta: editor.FieldMeta{
			Description: "Match files by size.",
		}},
	}
}

func (MatchFilter) Metadata() map[string]*metadata.Node {
	return map[string]*metadata.Node{
		"literal": {FieldMeta: editor.FieldMeta{
			Description: "Exact filename match (whole name including extension). Mutually exclusive with regex/glob.",
		}},
		"regex": {FieldMeta: editor.FieldMeta{
			Description: "RE2 regular expression matched against the filename. Mutually exclusive with literal/glob.",
			Example:     `regex: '^\d{4}-\d{2}-\d{2}_'`,
		}},
		"glob": {FieldMeta: editor.FieldMeta{
			Description: "Wildcard pattern matched against the filename. Mutually exclusive with literal/regex.",
			Example:     `glob: "screenshot_*"`,
		}},
		"case-sensitive": {FieldMeta: editor.FieldMeta{
			Description: "Whether literal/glob/regex matching is case-sensitive.",
			Default:     "false",
		}},
	}
}

func (AgeFilter) Metadata() map[string]*metadata.Node {
	return map[string]*metadata.Node{
		"min": {FieldMeta: editor.FieldMeta{
			Description: "Minimum time since the file was last modified.",
			Example:     "min: 24h",
		}},
		"max": {FieldMeta: editor.FieldMeta{
			Description: "Maximum time since the file was last modified.",
			Example:     "max: 720h",
		}},
	}
}

func (SizeFilter) Metadata() map[string]*metadata.Node {
	return map[string]*metadata.Node{
		"min": {FieldMeta: editor.FieldMeta{
			Description: "Minimum file size.",
			Example:     `min: "500KB"`,
		}},
		"max": {FieldMeta: editor.FieldMeta{
			Description: "Maximum file size.",
			Example:     `max: "20MB"`,
		}},
	}
}
