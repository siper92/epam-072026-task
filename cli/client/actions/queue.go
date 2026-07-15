package actions

import (
	"github.com/spf13/cobra"
)

func queueCommand(newClient ClientFactory, newPrinter PrinterFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "queue",
		Short: "join the matchmaking queue, pairs with the next waiting player",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := newClient()
			if err != nil {
				return err
			}
			game, err := c.QueueJoin(cmd.Context())
			if err != nil {
				return err
			}
			newPrinter().Joined(cmd, game)
			return nil
		},
	}
}
