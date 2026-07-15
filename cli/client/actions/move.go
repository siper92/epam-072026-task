package actions

import (
	"github.com/spf13/cobra"

	"ticTacSolved/task/game/utils"
)

func moveCommand(newClient ClientFactory, newPrinter PrinterFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "move <game-id> <cell>",
		Short: "make a move, cell is a1..c3, also accepts [game-id] <row> <col>",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient()
			if err != nil {
				return err
			}
			id, row, col, err := parseMoveArgs(args)
			if err != nil {
				return err
			}
			game, err := c.Move(cmd.Context(), id, row, col)
			if err != nil {
				return err
			}
			newPrinter().Game(cmd, game)
			return nil
		},
	}
}

func parseMoveArgs(args []string) (string, int, int, error) {
	if len(args) == 3 {
		row, col, err := utils.ParseCell(args[1], args[2])
		return args[0], row, col, err
	}
	if utils.IsCellName(args[1]) {
		row, col, err := utils.ParseCellName(args[1])
		return args[0], row, col, err
	}
	row, col, err := utils.ParseCell(args[0], args[1])
	return "", row, col, err
}
