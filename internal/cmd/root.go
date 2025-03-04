package cmd

import (
	"fmt"
	"gopaper/internal/config"
	"gopaper/internal/models"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
func RootCmd(g *models.Gopaper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gopaper",
		Short: "gopaper is a CLI tool for changing wallpapers",
		Long:  "gopaper is a CLI tool for changing wallpapers based on configurable categories",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ex, err := os.Executable()
			if err != nil {
				log.Fatalf("error getting executable: %v", err)
				return
			}

			options := []config.ViperOptions{
				config.WithConfigName("gopaper"),
				config.WithConfigType("yaml"),
				config.WithConfigPath("."), // This is being used to debug the application
				config.WithConfigPath(filepath.Dir(ex)),
				config.WithConfigPath(filepath.Join(filepath.Dir(ex), "conf")),
			}

			if err = config.InitConfig(g.Viper, options...); err != nil {
				log.Fatalf("error initializing configuration: %v", err)
				return
			}

			logger, err := config.ConfigureLogger(g.Viper)
			if err != nil {
				fmt.Printf("failed to configure logger: %v\n", err)
				return
			}

			g.Logger = logger

			if g.PersistentFlags == nil {
				g.Logger.Error("error configuring flags")
			}

			checkPersistentFlags(cmd, g, g.PersistentFlags, "output")
			checkPersistentFlags(cmd, g, g.PersistentFlags, "show-caller")
			checkPersistentFlags(cmd, g, g.PersistentFlags, "log-level")
		},
	}

	g.PersistentFlags = setPersistentFlags(cmd)

	bindPersistentFlag(cmd, g, "output")
	bindPersistentFlag(cmd, g, "log-level")
	bindPersistentFlag(cmd, g, "show-caller")

	cmd.AddCommand(ChangeCmd(g))
	cmd.AddCommand(BaseConfigCmd(g))

	return cmd
}

// setPersistentFlags sets the persistent flags for a Cobra command, which are flags
// that are available to the command and all of its subcommands.
func setPersistentFlags(cmd *cobra.Command) *models.PersistentFlags {
	return &models.PersistentFlags{
		ShowCaller: cmd.PersistentFlags().Bool("show-caller", false, "Show caller information"),
		LogLevel:   cmd.PersistentFlags().StringP("log-level", "l", "", "Specify the log level (trace, debug, info, warn/warning, error, fatal)"),
		Output:     cmd.PersistentFlags().StringP("output", "o", "", "Specify the output (console, log/file, both or none)"),
	}
}

// bindPersistentFlag links a CLI flag to a Viper key to enable configuration file support
func bindPersistentFlag(cmd *cobra.Command, g *models.Gopaper, flagName string) {
	// Bind the flag to a Viper key and handle any binding errors
	err := g.Viper.BindPFlag(fmt.Sprintf("configuration.%s", flagName), cmd.PersistentFlags().Lookup(flagName))
	if err != nil {
		g.Logger.Error("error binding flag", g.Logger.Args("flag", flagName, "error", err))
	}
}

// checkPersistentFlags ensures that the flags are set correctly, either from the command-line or from the Viper configuration
func checkPersistentFlags(cmd *cobra.Command, g *models.Gopaper, flags *models.PersistentFlags, flagName string) {
	// If the flag was not changed by the user, check Viper and set it if needed
	if !cmd.PersistentFlags().Changed(flagName) && g.Viper.IsSet(fmt.Sprintf("configuration.%s", flagName)) {
		switch flagName {
		case "output":
			*flags.Output = g.Viper.GetString(fmt.Sprintf("configuration.%s", flagName))
		case "log-level":
			*flags.LogLevel = g.Viper.GetString(fmt.Sprintf("configuration.%s", flagName))
		case "show-caller":
			*flags.ShowCaller = g.Viper.GetBool(fmt.Sprintf("configuration.%s", flagName))
		}
	}
}
