package game

import (
	"epam/task/pkg/errs"
)

func playerMark(state *GameState, playerID string) (Mark, error) {
	switch playerID {
	case state.PlayerX:
		return MarkX, nil
	case state.PlayerO:
		return MarkO, nil
	default:
		return MarkEmpty, errs.Newf(errs.CodeUnknownPlayer, "player %q is not part of game %q", playerID, state.ID)
	}
}

func turnMark(status state_machine.State) Mark {
	if status == state_machine.StateTurnO {
		return MarkO
	}
	return MarkX
}

func validateMove(state *GameState, mark Mark, row, col int) error {
	if turnMark(state.Status) != mark {
		return errs.Newf(errs.CodeOutOfTurn, "not player %q turn", string(mark))
	}
	if row < 0 || row >= GridSize || col < 0 || col >= GridSize {
		return errs.Newf(errs.CodeOutOfBounds, "cell (%d,%d) is outside the %dx%d grid", row, col, GridSize, GridSize)
	}
	if state.Grid[row][col] != MarkEmpty {
		return errs.Newf(errs.CodeCellOccupied, "cell (%d,%d) is already occupied", row, col)
	}
	return nil
}

func applyMove(state *GameState, mark Mark, row, col int) {
	state.Grid[row][col] = mark
	state.MoveCount++
}

func hasWin(grid Grid, mark Mark) bool {
	for i := 0; i < GridSize; i++ {
		if grid[i][0] == mark && grid[i][1] == mark && grid[i][2] == mark {
			return true
		}
		if grid[0][i] == mark && grid[1][i] == mark && grid[2][i] == mark {
			return true
		}
	}
	if grid[0][0] == mark && grid[1][1] == mark && grid[2][2] == mark {
		return true
	}
	if grid[0][2] == mark && grid[1][1] == mark && grid[2][0] == mark {
		return true
	}
	return false
}

func isFull(grid Grid) bool {
	for row := 0; row < GridSize; row++ {
		for col := 0; col < GridSize; col++ {
			if grid[row][col] == MarkEmpty {
				return false
			}
		}
	}
	return true
}
