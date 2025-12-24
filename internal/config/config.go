package config

import (
	"fmt"
	"log"

	"github.com/lucasassuncao/gopaper/internal/models"

	"github.com/spf13/viper"
)

// ViperOptions defines a function type for configuring Viper
type ViperOptions func(*viper.Viper)

// ConfigFileNotFoundError é um erro personalizado para quando o arquivo de config não é encontrado.
type ConfigFileNotFoundError struct {
	Err error
}

// Error implements the error interface for ConfigFileNotFoundError
func (e ConfigFileNotFoundError) Error() string {
	return fmt.Sprintf("config file not found: %v", e.Err)
}

// InitConfig inicializa o Viper
func InitConfig(v *viper.Viper, options ...ViperOptions) error {
	applyOptions(v, options...)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Embrulha o erro de arquivo não encontrado com o nosso erro personalizado
			return ConfigFileNotFoundError{Err: err}
		}
		return fmt.Errorf("não foi possível ler a config: %w", err)
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
func UnmarshalConfig(m *models.Gopaper) []*models.Categories {
	var categories []*models.Categories
	if err := m.Viper.UnmarshalKey("categories", &categories); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	return categories
}
