package cmd

import (
	"github.com/spf13/cobra"

	"ticTacSolved/task/pkg/config"
)

var rootCmd = &cobra.Command{
	Use:          "game-server",
	Short:        "Tic tac toe game service",
	SilenceUsage: true,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		config.LoadEnv()
	},
}

func Execute() error {
	return rootCmd.Execute()
}
