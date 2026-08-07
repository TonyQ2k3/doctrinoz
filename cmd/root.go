package cmd

import (
	"os"

	"github.com/TonyQ2k3/doctrinoz/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "doctrinoz",
	Short: "Agent-optimized wiki generator for code repositories.",
	Long: `Doctrinoz automatically generates and maintains AGENT.md that provides context for AI coding agents.
The tool uses git-aware incremental updates to keep documentations fresh without re-generating everything.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default: .doc.yaml)")
	rootCmd.PersistentFlags().String("provider", "openrouter", "AI model provider")
	rootCmd.PersistentFlags().String("model", "", "Model to use for content generation")
	rootCmd.PersistentFlags().String("api-key", "", "API Key for model provider")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")

	viper.BindPFlag("provider", rootCmd.PersistentFlags().Lookup("provider"))
	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	cobra.OnInitialize(func() { config.Init(configFile) })
}
