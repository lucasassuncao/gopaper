package config

import (
	"fmt"
	"os"
	"path/filepath"

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

// HistoryLimit returns the configured maximum number of history entries.
// A non-positive value tells history.Load to keep its own default.
func HistoryLimit(v *viper.Viper) int {
	return v.GetInt("configuration.history.limit")
}
