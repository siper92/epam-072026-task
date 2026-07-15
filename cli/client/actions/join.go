package actions

import (
	"github.com/spf13/cobra"

	"ticTacSolved/task/cli/client/internal"
)

func joinCommand(newClient ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join <game-id>",
		Short: "join a game and store its game token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient()
			if err != nil {
				return err
			}
			code, err := cmd.Flags().GetString("code")
			if err != nil {
				return err
			}
			game, err := c.JoinGame(cmd.Context(), args[0], code)
			if err != nil {
				return err
			}
			cmd.Print(internal.RenderGame(game))
			return nil
		},
	}
	cmd.Flags().String("code", "", "join code for a private game")
	return cmd
}
