package cmd

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/lucasassuncao/gopaper/internal/filters"
	"github.com/lucasassuncao/gopaper/internal/schedule"
	"github.com/lucasassuncao/gopaper/internal/weather"
	"github.com/lucasassuncao/yedit/editor"
)

// GopaperValidators is the rule set enforced by the edit command at
// validate/save time.
//
// Per-field constraints (required, allowed values, uniqueness) are declared
// once in the hint tree (models.Config.Metadata and friends) and enforced by
// the FromMetadata family — hints are the single source of field metadata.
// Only cross-field rules, which cannot live in per-field metadata, are
// declared here explicitly.
var GopaperValidators = []editor.Validator{
	// Enforce everything the metadata declares.
	editor.RequiredFromMetadata(),
	editor.OneOfFromMetadata(),
	editor.CountFromMetadata(),
	editor.UniqueFromMetadata(),

	// Category names must be unique across the list (presence comes from the
	// hints; NoDuplicates skips unnamed entries).
	editor.NoDuplicates("categories", "name"),

	// within a filter.match block, literal/regex/glob are mutually exclusive.
	editor.MutuallyExclusiveNested("categories.filter.match", "literal", "regex", "glob"),

	// age and size min/max pairs must be ordered.
	editor.CrossFieldOrderedNested("categories.filter.age", "min", "max"),
	editor.CrossFieldOrderedNested("categories.filter.size", "min", "max"),

	// Validate that filter.match.regex/glob compile and filter.size.min/max
	// parse, so a malformed filter is caught here instead of at wallpaper-
	// change time.
	editor.ValidatorFunc(func(in editor.ValidationInput) []editor.Violation {
		var doc struct {
			Categories []struct {
				Filter *struct {
					Match *struct {
						Regex string `yaml:"regex"`
						Glob  string `yaml:"glob"`
					} `yaml:"match"`
					Size *struct {
						Min string `yaml:"min"`
						Max string `yaml:"max"`
					} `yaml:"size"`
				} `yaml:"filter"`
			} `yaml:"categories"`
		}
		if err := yaml.Unmarshal(in.Raw, &doc); err != nil {
			return nil
		}
		var errs []editor.Violation
		for i, c := range doc.Categories {
			if c.Filter == nil {
				continue
			}
			if m := c.Filter.Match; m != nil {
				if m.Regex != "" {
					if _, err := regexp.Compile(m.Regex); err != nil {
						errs = append(errs, editor.Violation{
							Path:    fmt.Sprintf("categories[%d].filter.match.regex", i),
							Message: err.Error(),
						})
					}
				}
				if m.Glob != "" {
					if _, err := filepath.Match(m.Glob, ""); err != nil {
						errs = append(errs, editor.Violation{
							Path:    fmt.Sprintf("categories[%d].filter.match.glob", i),
							Message: err.Error(),
						})
					}
				}
			}
			if s := c.Filter.Size; s != nil {
				if s.Min != "" {
					if _, err := filters.ParseSize(s.Min); err != nil {
						errs = append(errs, editor.Violation{
							Path:    fmt.Sprintf("categories[%d].filter.size.min", i),
							Message: err.Error(),
						})
					}
				}
				if s.Max != "" {
					if _, err := filters.ParseSize(s.Max); err != nil {
						errs = append(errs, editor.Violation{
							Path:    fmt.Sprintf("categories[%d].filter.size.max", i),
							Message: err.Error(),
						})
					}
				}
			}
		}
		return errs
	}),

	// Category source/variants/wallhaven shape, and per-variant
	// hours/condition rules. A category has exactly one of: a plain source,
	// variants (source optional there, but required as the base directory
	// for any variant with a relative source), or a wallhaven block (which
	// requires a query, and an API key for sketchy/nsfw purity). Each
	// variant defines exactly one of hours/condition; a condition name must
	// exist in configuration.conditions.
	editor.ValidatorFunc(func(in editor.ValidationInput) []editor.Violation {
		var doc struct {
			Configuration struct {
				Wallhaven *struct {
					APIKey string `yaml:"api-key"`
				} `yaml:"wallhaven"`
				Conditions map[string]struct {
					Hours        string   `yaml:"hours"`
					Weather      []string `yaml:"weather"`
					WindSpeedMin *float64 `yaml:"wind-speed-min"`
					WindSpeedMax *float64 `yaml:"wind-speed-max"`
				} `yaml:"conditions"`
			} `yaml:"configuration"`
			Categories []struct {
				Source   string `yaml:"source"`
				Variants []struct {
					Source    string `yaml:"source"`
					Hours     string `yaml:"hours"`
					Condition string `yaml:"condition"`
				} `yaml:"variants"`
				Wallhaven *struct {
					Query  string `yaml:"query"`
					Purity string `yaml:"purity"`
				} `yaml:"wallhaven"`
			} `yaml:"categories"`
		}
		if err := yaml.Unmarshal(in.Raw, &doc); err != nil {
			return nil
		}
		hasAPIKey := doc.Configuration.Wallhaven != nil && doc.Configuration.Wallhaven.APIKey != ""
		var errs []editor.Violation
		for i, c := range doc.Categories {
			if c.Wallhaven != nil {
				if c.Source != "" || len(c.Variants) > 0 {
					errs = append(errs, editor.Violation{
						Path:    fmt.Sprintf("categories[%d].wallhaven", i),
						Message: "wallhaven is mutually exclusive with source/variants - define one or the other",
					})
				}
				if (c.Wallhaven.Purity == "sketchy" || c.Wallhaven.Purity == "nsfw") && !hasAPIKey {
					errs = append(errs, editor.Violation{
						Path:    fmt.Sprintf("categories[%d].wallhaven.purity", i),
						Message: fmt.Sprintf("%q requires configuration.wallhaven.api-key (anonymous searches only return sfw results)", c.Wallhaven.Purity),
					})
				}
				continue
			}
			if len(c.Variants) == 0 {
				if c.Source == "" {
					errs = append(errs, editor.Violation{
						Path:    fmt.Sprintf("categories[%d].source", i),
						Message: "define one of source, variants, or wallhaven",
					})
				}
				continue
			}
			for j, v := range c.Variants {
				if v.Source == "" {
					errs = append(errs, editor.Violation{
						Path:    fmt.Sprintf("categories[%d].variants[%d].source", i, j),
						Message: "required",
					})
				} else if !filepath.IsAbs(v.Source) && c.Source == "" {
					errs = append(errs, editor.Violation{
						Path:    fmt.Sprintf("categories[%d].variants[%d].source", i, j),
						Message: "relative source requires the category to define source (used as the base directory)",
					})
				}

				switch {
				case v.Hours != "" && v.Condition != "":
					errs = append(errs, editor.Violation{
						Path:    fmt.Sprintf("categories[%d].variants[%d].hours", i, j),
						Message: "hours and condition are mutually exclusive - define one or the other",
					})
				case v.Hours == "" && v.Condition == "":
					errs = append(errs, editor.Violation{
						Path:    fmt.Sprintf("categories[%d].variants[%d].hours", i, j),
						Message: `required - either "hours" (daily HH:MM-HH:MM window) or "condition" (name from configuration.conditions)`,
					})
				case v.Hours != "":
					if _, err := schedule.ParseWindow(v.Hours); err != nil {
						errs = append(errs, editor.Violation{
							Path:    fmt.Sprintf("categories[%d].variants[%d].hours", i, j),
							Message: err.Error(),
						})
					}
				case v.Condition != "":
					if _, ok := doc.Configuration.Conditions[v.Condition]; !ok {
						errs = append(errs, editor.Violation{
							Path:    fmt.Sprintf("categories[%d].variants[%d].condition", i, j),
							Message: fmt.Sprintf("unknown condition %q - not defined in configuration.conditions", v.Condition),
						})
					}
				}
			}
		}
		return errs
	}),

	// configuration.conditions shape: exactly one of hours / date-range /
	// weather-bucket (weather, wind-speed-*, temperature-*, which combine
	// with AND) per condition, known sky names, valid date-range, and
	// configuration.weather requiredness/validity.
	editor.ValidatorFunc(func(in editor.ValidationInput) []editor.Violation {
		var doc struct {
			Configuration struct {
				Weather *struct {
					Provider  string   `yaml:"provider"`
					Latitude  *float64 `yaml:"latitude"`
					Longitude *float64 `yaml:"longitude"`
					CacheTTL  string   `yaml:"cache-ttl"`
				} `yaml:"weather"`
				Conditions map[string]struct {
					Hours     string `yaml:"hours"`
					DateRange *struct {
						Start string `yaml:"start"`
						End   string `yaml:"end"`
					} `yaml:"date-range"`
					Weather        []string `yaml:"weather"`
					WindSpeedMin   *float64 `yaml:"wind-speed-min"`
					WindSpeedMax   *float64 `yaml:"wind-speed-max"`
					TemperatureMin *float64 `yaml:"temperature-min"`
					TemperatureMax *float64 `yaml:"temperature-max"`
				} `yaml:"conditions"`
			} `yaml:"configuration"`
		}
		if err := yaml.Unmarshal(in.Raw, &doc); err != nil {
			return nil
		}

		conditionNames := make([]string, 0, len(doc.Configuration.Conditions))
		for name := range doc.Configuration.Conditions {
			conditionNames = append(conditionNames, name)
		}
		sort.Strings(conditionNames)

		var errs []editor.Violation
		needsWeatherConfig := false
		for _, name := range conditionNames {
			cond := doc.Configuration.Conditions[name]
			hasHours := cond.Hours != ""
			hasDateRange := cond.DateRange != nil
			hasWeatherFields := len(cond.Weather) > 0 || cond.WindSpeedMin != nil || cond.WindSpeedMax != nil ||
				cond.TemperatureMin != nil || cond.TemperatureMax != nil

			groupCount := 0
			if hasHours {
				groupCount++
			}
			if hasDateRange {
				groupCount++
			}
			if hasWeatherFields {
				groupCount++
			}

			switch {
			case groupCount > 1:
				errs = append(errs, editor.Violation{
					Path:    fmt.Sprintf("configuration.conditions.%s", name),
					Message: "hours, date-range, and weather/wind-speed-*/temperature-* are mutually exclusive - define exactly one",
				})
			case groupCount == 0:
				errs = append(errs, editor.Violation{
					Path:    fmt.Sprintf("configuration.conditions.%s", name),
					Message: "define hours, date-range, or weather/wind-speed-*/temperature-*",
				})
			case hasDateRange:
				if cond.DateRange.Start == "" || cond.DateRange.End == "" {
					errs = append(errs, editor.Violation{
						Path:    fmt.Sprintf("configuration.conditions.%s.date-range", name),
						Message: `both start and end are required, in "MM-DD" format`,
					})
				} else if _, err := schedule.ParseDateRange(cond.DateRange.Start, cond.DateRange.End); err != nil {
					errs = append(errs, editor.Violation{
						Path:    fmt.Sprintf("configuration.conditions.%s.date-range", name),
						Message: err.Error(),
					})
				}
			case hasWeatherFields:
				needsWeatherConfig = true
				for _, sky := range cond.Weather {
					if !weather.IsValidSky(sky) {
						errs = append(errs, editor.Violation{
							Path:    fmt.Sprintf("configuration.conditions.%s.weather", name),
							Message: fmt.Sprintf("unknown weather category %q - use one of: %s", sky, strings.Join(weather.SkyNames(), ", ")),
						})
					}
				}
			}
		}

		if !needsWeatherConfig {
			return errs
		}

		w := doc.Configuration.Weather
		if w == nil {
			return append(errs, editor.Violation{
				Path:    "configuration.weather",
				Message: "required because a condition uses weather/wind-speed-min/wind-speed-max/temperature-min/temperature-max",
			})
		}
		if w.Provider != "open-meteo" {
			errs = append(errs, editor.Violation{
				Path:    "configuration.weather.provider",
				Message: `only "open-meteo" is supported`,
			})
		}
		if w.Latitude == nil || *w.Latitude < -90 || *w.Latitude > 90 {
			errs = append(errs, editor.Violation{
				Path:    "configuration.weather.latitude",
				Message: "required, must be between -90 and 90",
			})
		}
		if w.Longitude == nil || *w.Longitude < -180 || *w.Longitude > 180 {
			errs = append(errs, editor.Violation{
				Path:    "configuration.weather.longitude",
				Message: "required, must be between -180 and 180",
			})
		}
		if w.CacheTTL != "" {
			if _, err := time.ParseDuration(w.CacheTTL); err != nil {
				errs = append(errs, editor.Violation{
					Path:    "configuration.weather.cache-ttl",
					Message: err.Error(),
				})
			}
		}
		return errs
	}),

	// logging.file is required when logging.output is "log", "file", or "both".
	editor.ValidatorFunc(func(in editor.ValidationInput) []editor.Violation {
		var doc struct {
			Configuration struct {
				Logging struct {
					Output string `yaml:"output"`
					File   string `yaml:"file"`
				} `yaml:"logging"`
			} `yaml:"configuration"`
		}
		if err := yaml.Unmarshal(in.Raw, &doc); err != nil {
			return nil
		}
		switch doc.Configuration.Logging.Output {
		case "log", "file", "both":
		default:
			return nil
		}
		if doc.Configuration.Logging.File != "" {
			return nil
		}
		return []editor.Violation{{
			Path:    "configuration.logging.file",
			Message: fmt.Sprintf("required when output is %q", doc.Configuration.Logging.Output),
		}}
	}),
}
