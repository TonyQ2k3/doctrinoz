package cmd

import (
	"github.com/TonyQ2k3/doctrinoz/internal/config"
	"github.com/TonyQ2k3/doctrinoz/internal/wiki"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update [path]",
	Short: "Incrementally update a documentation based on git diff",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := "."
		if len(args) > 0 {
			repoPath = args[0]
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		return wiki.Update(repoPath, cfg)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
