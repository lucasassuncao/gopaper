package cmd

import (
	"fmt"
	"gopaper/internal/models"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var interactive bool

// BaseConfigCmd generates a base configuration file
func BaseConfigCmd(m *models.Gopaper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "baseconfig",
		Short: "Generates a base configuration file",
		Long: "Generates a base configuration file in the application directory with predefined categories.\n" +
			"This file can be customized to define category names, file extensions, source directories, and destination paths.\n" +
			"If the base configuration file already exists, it will not be overwritten.",
	}

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if m.Logger == nil {
			return fmt.Errorf("logger is not initialized")
		}

		m.Logger.Info("Creating a base configuration file")
		m.Logger.Debug("Using Configuration",
			m.Logger.Args("output", *m.PersistentFlags.Output),
			m.Logger.Args("show-caller", *m.PersistentFlags.ShowCaller),
			m.Logger.Args("log-level", *m.PersistentFlags.LogLevel),
			m.Logger.Args("log-file", m.Viper.GetString("configuration.log-file")),
			m.Logger.Args("config-file", m.Viper.ConfigFileUsed()),
		)

		return nil
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ex, err := os.Executable()
		if err != nil {
			m.Logger.Error("error getting executable", m.Logger.Args("error", err))
			return err
		}

		path := filepath.Join(filepath.Dir(ex), "conf", "base")

		err = createDirectory(path)
		if err != nil {
			m.Logger.Error("error creating directory for base config", m.Logger.Args("error", err))
		}

		var options = []models.ConfigOption{}

		if interactive {
			options = append(options, models.WithOutput())
			options = append(options, models.WithLogFile())
			options = append(options, models.WithLogLevel())
			options = append(options, models.WithShowCaller())

		}

		err = models.NewConfig(path, interactive, options...)
		if err != nil {
			m.Logger.Error("error creating base configuration file", m.Logger.Args("error", err))
		}

		m.Logger.Info("Base configuration file created", m.Logger.Args("path", path))

		return nil

	}

	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive mode for creating a base configuration file")

	return cmd
}
