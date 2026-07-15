package actions

import (
	"github.com/spf13/cobra"

	"ticTacSolved/task/cli/client/internal"
	"ticTacSolved/task/game/utils"
)

func moveCommand(newClient ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "move [game-id] <row> <col>",
		Short: "make a move, game id defaults to the stored current game",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient()
			if err != nil {
				return err
			}
			id := ""
			cellArgs := args
			if len(args) == 3 {
				id = args[0]
				cellArgs = args[1:]
			}

			row, col, err := utils.ParseCell(cellArgs[0], cellArgs[1])
			if err != nil {
				return err
			}
			game, err := c.Move(cmd.Context(), id, row, col)
			if err != nil {
				return err
			}
			cmd.Print(internal.RenderGame(game))
			return nil
		},
	}
}
