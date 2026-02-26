package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gyurisc/adocli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  "Get, set, and list ado CLI configuration values.",
}

// --- ado config set ---

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value. Valid keys:
  organization   Azure DevOps organization name
  project        Default project name
  output_format  Default output format (table, json, plain)`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch key {
	case "organization":
		cfg.Organization = value
	case "project":
		cfg.Project = value
	case "output_format":
		if value != "table" && value != "json" && value != "plain" {
			return fmt.Errorf("invalid output_format %q (must be table, json, or plain)", value)
		}
		cfg.OutputFormat = value
	default:
		return fmt.Errorf("unknown config key %q (valid: organization, project, output_format)", key)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Set %s = %s\n", key, value)
	return nil
}

// --- ado config get ---

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var value string
	switch key {
	case "organization":
		value = cfg.Organization
	case "project":
		value = cfg.Project
	case "output_format":
		value = cfg.OutputFormat
	default:
		return fmt.Errorf("unknown config key %q (valid: organization, project, output_format)", key)
	}

	fmt.Println(value)
	return nil
}

// --- ado config list ---

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	RunE:  runConfigList,
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch OutputFormat() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(cfg)
	default:
		fmt.Printf("organization  = %s\n", cfg.Organization)
		fmt.Printf("project       = %s\n", cfg.Project)
		fmt.Printf("output_format = %s\n", cfg.OutputFormat)
		path, err := config.Path()
		if err == nil {
			fmt.Printf("\nConfig file: %s\n", path)
		}
	}
	return nil
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)

	rootCmd.AddCommand(configCmd)
}
