package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/lucasassuncao/gopaper/internal/models"
	"github.com/lucasassuncao/yedit/presets"
)

var GopaperBlockPresets = presets.Combine(
	presets.ForField("configuration", configurationPresetsMap()),
	presets.ForField("categories", categoriesPresetsMap()),
)

// GopaperDocPresets is a whole-document preset source for the root template
// picker (ctrl+p). Each entry is a full gopaper.yaml ready to use as a
// starting point.
var GopaperDocPresets presets.Source = buildDocPresets()

// docPresetSource implements presets.Source for whole-document templates.
// PresetYAML("", name) returns the full YAML for the named template.
type docPresetSource struct {
	names []string
	yamls map[string]string
}

func (s *docPresetSource) ListFields() []string { return []string{""} }
func (s *docPresetSource) ListPresets(field string) []string {
	if field != "" {
		return nil
	}
	return s.names
}
func (s *docPresetSource) PresetYAML(field, name string) (string, error) {
	if field != "" {
		return "", fmt.Errorf("docPresetSource: unknown field %q", field)
	}
	y, ok := s.yamls[name]
	if !ok {
		return "", fmt.Errorf("docPresetSource: unknown preset %q", name)
	}
	return y, nil
}

func buildDocPresets() *docPresetSource {
	docs := docPresetsMap()

	names := make([]string, 0, len(docs))
	for name := range docs {
		names = append(names, name)
	}
	sort.Strings(names)

	yamls := make(map[string]string, len(docs))
	for _, name := range names {
		raw, err := yaml.Marshal(docs[name])
		if err != nil {
			continue
		}
		yamls[name] = string(raw)
	}
	return &docPresetSource{names: names, yamls: yamls}
}

func docPresetsMap() map[string]*models.Config {
	walls := "~/Pictures/Walls"

	return map[string]*models.Config{
		// single category, console logging: the smallest working config
		"single-category": {
			Configuration: models.Configuration{
				Logging: models.Logging{Output: "console", Level: "info"},
			},
			Categories: []models.Categories{
				{
					Name:    "Custom Selection",
					Source:  filepath.Join(walls, "CustomSelection"),
					Mode:    "crop",
					Enabled: true,
				},
			},
		},
		// multiple categories with mixed modes and only one enabled
		"multi-category": {
			Configuration: models.Configuration{
				Logging: models.Logging{Output: "console", Level: "info"},
			},
			Categories: []models.Categories{
				{
					Name:    "Wallhaven",
					Source:  filepath.Join(walls, "Wallhaven"),
					Mode:    "crop",
					Enabled: true,
				},
				{
					Name:    "Nature",
					Source:  filepath.Join(walls, "Nature"),
					Mode:    "fit",
					Enabled: false,
				},
				{
					Name:    "Minimal",
					Source:  filepath.Join(walls, "Minimal"),
					Mode:    "center",
					Enabled: false,
				},
			},
		},
		// file logging: every run is appended to a log file instead of the console
		"file-logging": {
			Configuration: models.Configuration{
				Logging: models.Logging{Output: "file", File: "~/.gopaper/logs/gopaper.log", Level: "info"},
			},
			Categories: []models.Categories{
				{
					Name:    "Custom Selection",
					Source:  filepath.Join(walls, "CustomSelection"),
					Mode:    "crop",
					Enabled: true,
				},
			},
		},
		// both: console output for interactive use plus a persistent file log
		"console-and-file": {
			Configuration: models.Configuration{
				Logging: models.Logging{Output: "both", File: "~/.gopaper/logs/gopaper.log", Level: "info", ShowCaller: true},
			},
			Categories: []models.Categories{
				{
					Name:    "Custom Selection",
					Source:  filepath.Join(walls, "CustomSelection"),
					Mode:    "crop",
					Enabled: true,
				},
			},
		},
		// history-limited: keeps only the last 10 wallpapers for prev/next
		"history-limited": {
			Configuration: models.Configuration{
				Logging: models.Logging{Output: "console", Level: "info"},
				History: models.History{Limit: 10},
			},
			Categories: []models.Categories{
				{
					Name:    "Custom Selection",
					Source:  filepath.Join(walls, "CustomSelection"),
					Mode:    "crop",
					Enabled: true,
				},
			},
		},
	}
}

func configurationPresetsMap() map[string]*models.Configuration {
	cfg := func(output, level string, showCaller bool) *models.Configuration {
		return &models.Configuration{
			Logging: models.Logging{
				Output:     output,
				File:       "~/.gopaper/logs/gopaper.log",
				Level:      level,
				ShowCaller: showCaller,
			},
		}
	}

	return map[string]*models.Configuration{
		"console-info":  cfg("console", "info", false),
		"console-debug": cfg("console", "debug", true),
		"console-trace": cfg("console", "trace", true),
		"file":          cfg("file", "warn", false),
		"both":          cfg("both", "info", false),
		"none":          {Logging: models.Logging{Output: "none", Level: "info"}},
	}
}

func ConfigurationPreset(name string) *models.Configuration {
	return configurationPresetsMap()[name]
}

func ListOfConfigurationPresets() []string {
	field := "configuration"
	return presets.ForField(field, configurationPresetsMap()).ListPresets(field)
}

func categoriesPresetsMap() map[string][]models.Categories {
	walls := "~/Pictures/Walls"

	mode := func(name, source, m string) models.Categories {
		return models.Categories{
			Name:    name,
			Source:  source,
			Mode:    m,
			Enabled: true,
		}
	}

	return map[string][]models.Categories{
		// mode.crop: fills the screen, cropping edges that don't fit the aspect ratio
		"with-mode-crop": {
			mode("Wallpapers", filepath.Join(walls, "Wallpapers"), "crop"),
		},
		// mode.tile: repeats the image at its native size to fill the screen
		"with-mode-tile": {
			mode("Patterns", filepath.Join(walls, "Patterns"), "tile"),
		},
		// mode.stretch: stretches the image to fill the screen, ignoring aspect ratio
		"with-mode-stretch": {
			mode("Banners", filepath.Join(walls, "Banners"), "stretch"),
		},
		// mode.span: stretches the image across all connected monitors
		"with-mode-span": {
			mode("Panoramas", filepath.Join(walls, "Panoramas"), "span"),
		},
		// mode.fit: scales the image to fit within the screen without cropping
		"with-mode-fit": {
			mode("Photography", filepath.Join(walls, "Photography"), "fit"),
		},
		// mode.center: centers the image at its native size, no scaling
		"with-mode-center": {
			mode("Icons", filepath.Join(walls, "Icons"), "center"),
		},
		// filter.match.glob: only files whose name matches a wildcard pattern
		"with-filter-match-glob": {
			{
				Name:    "Screenshots",
				Source:  filepath.Join(walls, "Screenshots"),
				Mode:    "fit",
				Enabled: true,
				Filter:  &models.Filter{Match: &models.MatchFilter{Glob: "screenshot_*"}},
			},
		},
		// filter.match.regex: only files whose name matches an RE2 pattern
		"with-filter-match-regex": {
			{
				Name:    "Dated",
				Source:  filepath.Join(walls, "Dated"),
				Mode:    "crop",
				Enabled: true,
				Filter:  &models.Filter{Match: &models.MatchFilter{Regex: `^\d{4}-\d{2}-\d{2}_`}},
			},
		},
		// filter.age: only files modified within a recent window
		"with-filter-age": {
			{
				Name:    "Recent",
				Source:  filepath.Join(walls, "Recent"),
				Mode:    "crop",
				Enabled: true,
				Filter:  &models.Filter{Age: &models.AgeFilter{Max: 30 * 24 * time.Hour}},
			},
		},
		// filter.size: only files within a byte-size range
		"with-filter-size": {
			{
				Name:    "HighRes",
				Source:  filepath.Join(walls, "HighRes"),
				Mode:    "crop",
				Enabled: true,
				Filter:  &models.Filter{Size: &models.SizeFilter{Min: "2MB"}},
			},
		},
	}
}

func CategoriesPreset(name string) []models.Categories {
	return categoriesPresetsMap()[name]
}

func ListOfCategoriesPresets() []string {
	field := "categories"
	return presets.ForField(field, categoriesPresetsMap()).ListPresets(field)
}
