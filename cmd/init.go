package cmd

import (
	"github.com/TonyQ2k3/doctrinoz/internal/config"
	"github.com/TonyQ2k3/doctrinoz/internal/wiki"
	"github.com/spf13/cobra"
)

// Argument values
var path string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialise documentation for a code repository",
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
		return wiki.Init(repoPath, cfg)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&path, "path", "", "Path of code repository to scan and generate docs.")
}
