/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, _ := cmd.Flags().GetString("provider")
		model, _ := cmd.Flags().GetString("model")
		apiKey, _ := cmd.Flags().GetString("api-key")

		providerChanged := cmd.Flags().Changed("provider")
		modelChanged := cmd.Flags().Changed("model")
		apiKeyChanged := cmd.Flags().Changed("api-key")

		if apiKeyChanged && provider == "" {
			return fmt.Errorf("--provider is required when setting --api-key")
		}

		if apiKeyChanged {
			if err := config.SetKey(provider, apiKey); err != nil {
				return err
			}
		}
		// Write provider config but not API key into .doc.yaml
		// --provider and --api-key being set at the same time only update API key, not the provider
		if providerChanged && !apiKeyChanged {
			if err := config.SetProvider(provider); err != nil {
				return err
			}
		}
		if modelChanged {
			if err := config.SetModel(model); err != nil {
				return err
			}
		}
		return nil
	},
}

var configSetProviderCmd = &cobra.Command{
	Use:        "provider <name>",
	Short:      "Set your model provider",
	Deprecated: "Use 'config set --provider <provider-name>' instead.",
	Args:       cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.SetProvider(args[0])
	},
}

var configSetModelCmd = &cobra.Command{
	Use:        "model <name>",
	Short:      "Set your model used for doc generation",
	Deprecated: "Use 'config set --model <model-name>' instead.",
	Args:       cobra.ExactArgs(1),
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
