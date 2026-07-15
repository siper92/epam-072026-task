package actions

import (
	"github.com/spf13/cobra"

	"ticTacSolved/task/cli/client/internal"
)

func createCommand(newClient ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a new game and store its game token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := newClient()
			if err != nil {
				return err
			}
			private, err := cmd.Flags().GetBool("private")
			if err != nil {
				return err
			}
			game, err := c.CreateGame(cmd.Context(), !private)
			if err != nil {
				return err
			}
			cmd.Print(internal.RenderGame(game))
			return nil
		},
	}
	cmd.Flags().Bool("private", false, "create a private game with a join code")
	return cmd
}
