/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/TonyQ2k3/doctrinoz/internal/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configurations",
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cfg.Print()
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a configuration value",
}

var configSetProviderCmd = &cobra.Command{
	Use:   "provider <name>",
	Short: "Set your model provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.SetProvider(args[0])
	},
}

var configSetModelCmd = &cobra.Command{
	Use:   "model <name>",
	Short: "Set your model used for doc generation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.SetModel(args[0])
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configSetCmd)

	configSetCmd.AddCommand(configSetProviderCmd)
	configSetCmd.AddCommand(configSetModelCmd)

	rootCmd.AddCommand(configCmd)
}
