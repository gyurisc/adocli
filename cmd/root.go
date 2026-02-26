// Package cmd implements all CLI commands for ado.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	jsonOutput  bool
	plainOutput bool
	appVersion  string
)

// SetVersion sets the application version (called from main with ldflags value).
func SetVersion(v string) {
	appVersion = v
	rootCmd.Version = v
}

// OutputFormat returns the current output format based on flags.
// Priority: --json > --plain > config > "table" (default).
func OutputFormat() string {
	if jsonOutput {
		return "json"
	}
	if plainOutput {
		return "plain"
	}
	if f := viper.GetString("output_format"); f != "" {
		return f
	}
	return "table"
}

var rootCmd = &cobra.Command{
	Use:   "ado",
	Short: "A fast, script-friendly CLI for Azure DevOps",
	Long: `ado is a command-line tool for interacting with Azure DevOps services.
It supports work items, pull requests, pipelines, repos, and more.

Configure with: ado auth login
Config file:    ~/.config/ado/config.json`,
	SilenceUsage: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&plainOutput, "plain", false, "Output in plain text (no colors, no borders)")

	rootCmd.SetVersionTemplate(fmt.Sprintf("ado version %s\n", appVersion))
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("$HOME/.config/ado")
	viper.SetEnvPrefix("ADO")
	viper.AutomaticEnv()

	// Silently ignore missing config file; it's created on first use.
	_ = viper.ReadInConfig()
}
