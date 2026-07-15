package actions

import (
	"github.com/spf13/cobra"

	"ticTacSolved/task/pkg/errs"
)

func showCommand(newClient ClientFactory, newPrinter PrinterFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "show [game-id]",
		Short: "show a game, defaults to the stored current game",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient()
			if err != nil {
				return err
			}
			id := ""
			if len(args) == 1 {
				id = args[0]
			} else {
				data, err := c.Session()
				if err != nil {
					return err
				}
				id = data.GameID
			}
			if id == "" {
				return errs.New(errs.CodeInvalidInput, "game id is required")
			}
			game, err := c.GetGame(cmd.Context(), id)
			if err != nil {
				return err
			}
			newPrinter().Game(cmd, game)
			return nil
		},
	}
}
