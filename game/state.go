package game

import (
	"epam/task/game/state_machine"
	"epam/task/pkg/errs"
	"epam/task/pkg/util"
)

type Mark rune

const (
	MarkEmpty Mark = '_'
	MarkX     Mark = 'X'
	MarkO     Mark = 'O'
)

const GridSize = 3

type Grid [GridSize][GridSize]Mark

func NewGrid() Grid {
	var grid Grid
	for row := 0; row < GridSize; row++ {
		for col := 0; col < GridSize; col++ {
			grid[row][col] = MarkEmpty
		}
	}
	return grid
}

func (g Grid) Encode() string {
	var cells [GridSize][GridSize]rune
	for row := range g {
		for col := range g[row] {
			cells[row][col] = rune(g[row][col])
		}
	}
	return util.EncodeGrid(cells)
}

func ParseGrid(encoded string) (Grid, error) {
	cells, err := util.DecodeGrid(encoded)
	if err != nil {
		return Grid{}, err
	}
	var grid Grid
	for row := range cells {
		for col := range cells[row] {
			mark := Mark(cells[row][col])
			if mark != MarkEmpty && mark != MarkX && mark != MarkO {
				return Grid{}, errs.Newf(errs.CodeInvalidInput, "invalid mark %q in encoded grid", string(mark))
			}
			grid[row][col] = mark
		}
	}
	return grid, nil
}

type GameState struct {
	ID        string
	PlayerX   string
	PlayerO   string
	Grid      Grid
	Status    state_machine.State
	WinnerID  string
	MoveCount int
}

func (s GameState) Finished() bool {
	return state_machine.IsTerminal(s.Status)
}
