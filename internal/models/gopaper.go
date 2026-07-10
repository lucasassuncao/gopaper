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
	Logging    Logging              `yaml:"logging" mapstructure:"logging"`
	History    History              `yaml:"history" mapstructure:"history"`
	Transition string               `yaml:"transition,omitempty" mapstructure:"transition"`
	Weather    *WeatherConfig       `yaml:"weather,omitempty" mapstructure:"weather"`
	Conditions map[string]Condition `yaml:"conditions,omitempty" mapstructure:"conditions"`
}

// WeatherConfig configures the weather data source used by
// weather-based conditions.
type WeatherConfig struct {
	Provider  string  `yaml:"provider" mapstructure:"provider"`
	Latitude  float64 `yaml:"latitude" mapstructure:"latitude"`
	Longitude float64 `yaml:"longitude" mapstructure:"longitude"`
	CacheTTL  string  `yaml:"cache-ttl,omitempty" mapstructure:"cache-ttl"`
}

// Condition is a named, reusable rule a variant can reference by name
// instead of declaring hours inline. It holds via exactly one of: hours,
// date-range, or the weather bucket (weather/wind-speed-*/temperature-*,
// which combine with AND). Priority breaks ties when multiple variants'
// conditions hold at the same time (higher wins); it defaults to 0.
type Condition struct {
	Hours          string     `yaml:"hours,omitempty" mapstructure:"hours"`
	DateRange      *DateRange `yaml:"date-range,omitempty" mapstructure:"date-range"`
	Weather        []string   `yaml:"weather,omitempty" mapstructure:"weather"`
	WindSpeedMin   *float64   `yaml:"wind-speed-min,omitempty" mapstructure:"wind-speed-min"`
	WindSpeedMax   *float64   `yaml:"wind-speed-max,omitempty" mapstructure:"wind-speed-max"`
	TemperatureMin *float64   `yaml:"temperature-min,omitempty" mapstructure:"temperature-min"`
	TemperatureMax *float64   `yaml:"temperature-max,omitempty" mapstructure:"temperature-max"`
	Priority       int        `yaml:"priority,omitempty" mapstructure:"priority"`
}

// DateRange is an inclusive calendar span (month/day only, no year); it
// wraps into the next year when Start orders after End (e.g. start
// "12-21", end "03-20" spans New Year's Eve), the same way Hours wraps
// past midnight.
type DateRange struct {
	Start string `yaml:"start" mapstructure:"start"`
	End   string `yaml:"end" mapstructure:"end"`
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
		"transition": {FieldMeta: editor.FieldMeta{
			Description: "Visual effect used when the wallpaper changes. \"fade\" uses the native Windows crossfade (falls back to an instant change on other platforms or if unavailable); \"none\" always changes instantly.",
			OneOf:       []string{"fade", "none"},
			Default:     "fade",
		}},
		"weather": {FieldMeta: editor.FieldMeta{
			Description: "Location and provider settings for weather-based conditions. Required if any entry in configuration.conditions uses weather, wind-speed-min, or wind-speed-max.",
		}},
		"conditions": {FieldMeta: editor.FieldMeta{
			Description: "Named, reusable conditions (time-of-day or weather) referenced by categories[].variants[].condition.",
		}},
	}
}

func (WeatherConfig) Metadata() map[string]*metadata.Node {
	return map[string]*metadata.Node{
		"provider": {FieldMeta: editor.FieldMeta{
			Description: "Weather data provider.",
			Required:    true,
			OneOf:       []string{"open-meteo"},
			Default:     "open-meteo",
		}},
		"latitude": {FieldMeta: editor.FieldMeta{
			Description: "Latitude of the location used for weather conditions, in decimal degrees.",
			Required:    true,
			Min:         "-90",
			Max:         "90",
		}},
		"longitude": {FieldMeta: editor.FieldMeta{
			Description: "Longitude of the location used for weather conditions, in decimal degrees.",
			Required:    true,
			Min:         "-180",
			Max:         "180",
		}},
		"cache-ttl": {FieldMeta: editor.FieldMeta{
			Description: `How long a fetched weather snapshot is reused before refetching, as a Go duration (e.g. "15m").`,
			Default:     "15m",
		}},
	}
}

func (Condition) Metadata() map[string]*metadata.Node {
	return map[string]*metadata.Node{
		"hours": {FieldMeta: editor.FieldMeta{
			Description: "Daily time window in 24h HH:MM-HH:MM format, both ends inclusive, may cross midnight. Mutually exclusive with date-range and weather/wind-speed-*/temperature-*.",
			Example:     `hours: "18:00-05:59"`,
		}},
		"date-range": {FieldMeta: editor.FieldMeta{
			Description: "Calendar date span (month/day only, no year), both ends inclusive; wraps into the next year when start orders after end. Mutually exclusive with hours and weather/wind-speed-*/temperature-*.",
		}},
		"weather": {FieldMeta: editor.FieldMeta{
			Description: "Sky conditions that satisfy this condition: one or more of clear, cloudy, fog, drizzle, rain, snow, thunderstorm. Combinable with wind-speed-*/temperature-* (AND); mutually exclusive with hours/date-range.",
		}},
		"wind-speed-min": {FieldMeta: editor.FieldMeta{
			Description: "Minimum current wind speed, in km/h, for this condition to hold.",
		}},
		"wind-speed-max": {FieldMeta: editor.FieldMeta{
			Description: "Maximum current wind speed, in km/h, for this condition to hold.",
		}},
		"temperature-min": {FieldMeta: editor.FieldMeta{
			Description: "Minimum current temperature, in Celsius, for this condition to hold.",
		}},
		"temperature-max": {FieldMeta: editor.FieldMeta{
			Description: "Maximum current temperature, in Celsius, for this condition to hold.",
		}},
		"priority": {FieldMeta: editor.FieldMeta{
			Description: "Tie-breaker when multiple variants' conditions hold at once; the highest priority wins. Default 0.",
			Default:     "0",
		}},
	}
}

func (DateRange) Metadata() map[string]*metadata.Node {
	return map[string]*metadata.Node{
		"start": {FieldMeta: editor.FieldMeta{
			Description: `Start of the date span, in "MM-DD" format (zero-padded).`,
			Required:    true,
			Example:     `start: "12-21"`,
		}},
		"end": {FieldMeta: editor.FieldMeta{
			Description: `End of the date span, in "MM-DD" format (zero-padded), inclusive.`,
			Required:    true,
			Example:     `end: "03-20"`,
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
			Description: "Directory containing the wallpaper images for this category, used directly when there are no variants, or as the base directory for variants with relative source paths.",
		}},
		"variants": {FieldMeta: editor.FieldMeta{
			Description: "Time-conditioned renditions of this category (e.g. day/night). The first variant whose hours window contains the current time provides the source; if none matches, the category is skipped for that run. Mutually exclusive with source.",
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
	Name     string    `yaml:"name" mapstructure:"name"`
	Source   string    `yaml:"source,omitempty" mapstructure:"source"`
	Mode     string    `yaml:"mode" mapstructure:"mode"`
	Enabled  bool      `yaml:"enabled" mapstructure:"enabled"`
	Filter   *Filter   `yaml:"filter,omitempty" mapstructure:"filter"`
	Variants []Variant `yaml:"variants,omitempty" mapstructure:"variants"`
}

// Variant is one conditioned rendition of a category's image collection
// (macOS dynamic wallpaper style). Among variants whose condition
// currently holds, the one with the highest priority provides the
// category's source.
type Variant struct {
	Source    string `yaml:"source" mapstructure:"source"`
	Hours     string `yaml:"hours,omitempty" mapstructure:"hours"`
	Condition string `yaml:"condition,omitempty" mapstructure:"condition"`
}

func (Variant) Metadata() map[string]*metadata.Node {
	return map[string]*metadata.Node{
		"source": {FieldMeta: editor.FieldMeta{
			Description: `Directory containing this variant's wallpaper images. Absolute paths are used as-is; relative paths (e.g. "./day") are resolved against the category's source.`,
		}},
		"hours": {FieldMeta: editor.FieldMeta{
			Description: "Daily time window in which this variant is active, in 24h HH:MM-HH:MM format, both ends inclusive. May cross midnight. Mutually exclusive with condition.",
			Example:     `hours: "18:00-05:59"`,
		}},
		"condition": {FieldMeta: editor.FieldMeta{
			Description: "Name of a condition declared in configuration.conditions. Mutually exclusive with hours.",
		}},
	}
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
