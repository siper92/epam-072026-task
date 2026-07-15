package actions

import (
	"github.com/spf13/cobra"
)

func stateCommand(newClient ClientFactory, newPrinter PrinterFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "state <game-id>",
		Short: "print the current state of a game",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient()
			if err != nil {
				return err
			}
			game, err := c.GetGame(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			newPrinter().Game(cmd, game)
			return nil
		},
	}
}
