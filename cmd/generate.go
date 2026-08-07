package cmd

import (
	"github.com/TonyQ2k3/doctrinoz/internal/config"
	"github.com/TonyQ2k3/doctrinoz/internal/wiki"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [path]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := "."
		if len(args) > 0 {
			repoPath = args[0]
		}
		force, _ := cmd.Flags().GetBool("force")
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		return wiki.Generate(repoPath, cfg, force)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().Bool("force", false, "Force generate all pages even if unchanged")
}
