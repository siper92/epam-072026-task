package actions

import (
	"github.com/spf13/cobra"
)

func listCommand(newClient ClientFactory, newPrinter PrinterFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list public games waiting for players",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := newClient()
			if err != nil {
				return err
			}
			games, err := c.WaitingGames(cmd.Context())
			if err != nil {
				return err
			}
			newPrinter().Games(cmd, games)
			return nil
		},
	}
}
