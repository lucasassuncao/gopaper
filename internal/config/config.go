package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucasassuncao/gopaper/internal/history"
	"github.com/lucasassuncao/gopaper/internal/models"

	"github.com/spf13/viper"
)

// ViperOptions defines a function type for configuring Viper
type ViperOptions func(*viper.Viper)

// ConfigFileNotFoundError is a custom error for when the config file is not found.
type ConfigFileNotFoundError struct {
	Err error
}

// Error implements the error interface for ConfigFileNotFoundError
func (e ConfigFileNotFoundError) Error() string {
	return fmt.Sprintf("config file not found: %v", e.Err)
}

// InitConfig initializes Viper with the provided options and reads the config file.
func InitConfig(v *viper.Viper, options ...ViperOptions) error {
	applyOptions(v, options...)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return ConfigFileNotFoundError{Err: err}
		}
		return fmt.Errorf("could not read config: %w", err)
	}
	return nil
}

// applyOptions applies the options to the viper instance
func applyOptions(v *viper.Viper, options ...ViperOptions) {
	for _, option := range options {
		option(v)
	}
}

// WithConfigName sets the name of the config file
func WithConfigName(name string) ViperOptions {
	return func(v *viper.Viper) {
		v.SetConfigName(name)
	}
}

// WithConfigType sets the type of the config file
func WithConfigType(configType string) ViperOptions {
	return func(v *viper.Viper) {
		v.SetConfigType(configType)
	}
}

// WithConfigPath sets the path of the config file
func WithConfigPath(path string) ViperOptions {
	return func(v *viper.Viper) {
		v.AddConfigPath(path)
	}
}

// UnmarshalConfig unmarshals the config file into a struct
func UnmarshalConfig(m *models.Gopaper) ([]*models.Categories, error) {
	var categories []*models.Categories
	if err := m.Viper.UnmarshalKey("categories", &categories); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return categories, nil
}

// LoadDefault loads gopaper.yaml from the standard search locations: next to
// the executable, then its conf subdirectory. Returns
// ConfigFileNotFoundError if none exists.
func LoadDefault(v *viper.Viper) error {
	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable: %w", err)
	}

	return InitConfig(v,
		WithConfigName("gopaper"),
		WithConfigType("yaml"),
		WithConfigPath(filepath.Dir(ex)),
		WithConfigPath(filepath.Join(filepath.Dir(ex), "conf")),
	)
}

// TransitionEnabledForCategory reports whether wallpaper changes for a
// category should use the fade transition. A non-empty categoryTransition
// (categories[].behavior.transition) overrides the configuration-level
// behavior.transition; both default to fade when unset.
func TransitionEnabledForCategory(v *viper.Viper, categoryTransition string) bool {
	t := categoryTransition
	if t == "" {
		t = v.GetString("configuration.behavior.transition")
	}
	return t != "none"
}

// MultiMonitorMode returns the configured default multi-monitor behavior:
// "per-monitor" when explicitly set, otherwise "same".
func MultiMonitorMode(v *viper.Viper) string {
	if v.GetString("configuration.behavior.multi-monitor") == "per-monitor" {
		return "per-monitor"
	}
	return "same"
}

// MultiMonitorModeForCategory resolves the effective multi-monitor mode for
// a category: its own behavior.multi-monitor when set, otherwise the
// configuration-level default.
func MultiMonitorModeForCategory(v *viper.Viper, categoryMode string) string {
	switch categoryMode {
	case "same", "per-monitor":
		return categoryMode
	}
	return MultiMonitorMode(v)
}

// LoadWallhavenAPIKey returns configuration.wallhaven.api-key, or "" when unset.
func LoadWallhavenAPIKey(v *viper.Viper) string {
	return v.GetString("configuration.wallhaven.api-key")
}

// WallhavenCacheDir resolves a category's Wallhaven cache directory: the
// category's own cache override when set (tilde-expanded), otherwise a
// wallhaven-cache/<slug> subdirectory next to the history file.
func WallhavenCacheDir(v *viper.Viper, categoryName, override string) (string, error) {
	if override != "" {
		return ExpandTilde(override), nil
	}
	histPath, err := HistoryPath(v)
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(histPath), "wallhaven-cache", slugify(categoryName)), nil
}

// slugify lowercases name and collapses every non-alphanumeric run into a
// single "-", trimmed at both ends, for use as a directory name.
func slugify(name string) string {
	var b strings.Builder
	lastDash := true // suppress a leading dash
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// HistoryEnabled reports whether wallpaper changes should be recorded to
// history. Defaults to true when configuration.history.enabled is not set.
func HistoryEnabled(v *viper.Viper) bool {
	if !v.IsSet("configuration.history.enabled") {
		return true
	}
	return v.GetBool("configuration.history.enabled")
}

// HistoryPath returns the configured history file path, falling back to
// history.DefaultPath() when configuration.history.file is not set.
func HistoryPath(v *viper.Viper) (string, error) {
	if file := v.GetString("configuration.history.file"); file != "" {
		return ExpandTilde(file), nil
	}
	return history.DefaultPath()
}

// LoadConditions returns the named conditions declared in
// configuration.conditions, keyed by name. Returns an empty (non-nil) map
// when the section is absent.
func LoadConditions(v *viper.Viper) (map[string]models.Condition, error) {
	conditions := map[string]models.Condition{}
	if err := v.UnmarshalKey("configuration.conditions", &conditions); err != nil {
		return nil, fmt.Errorf("unable to decode configuration.conditions: %w", err)
	}
	return conditions, nil
}

// LoadWeatherConfig returns the configuration.weather section, or nil when
// it is not set.
func LoadWeatherConfig(v *viper.Viper) (*models.WeatherConfig, error) {
	if !v.IsSet("configuration.weather") {
		return nil, nil
	}
	var wc models.WeatherConfig
	if err := v.UnmarshalKey("configuration.weather", &wc); err != nil {
		return nil, fmt.Errorf("unable to decode configuration.weather: %w", err)
	}
	return &wc, nil
}

// WeatherCachePath returns the path to the cached weather snapshot, in the
// same directory as the history file.
func WeatherCachePath() (string, error) {
	histPath, err := history.DefaultPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(histPath), "weather-cache.json"), nil
}

// HistoryLimit returns the configured maximum number of history entries.
// A non-positive value tells history.Load to keep its own default.
func HistoryLimit(v *viper.Viper) int {
	return v.GetInt("configuration.history.limit")
}
