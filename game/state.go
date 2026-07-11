package game

import (
	"epam/task/game/statemachine"
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

type GameState struct {
	ID        string
	PlayerX   string
	PlayerO   string
	Grid      Grid
	Status    statemachine.State
	WinnerID  string
	MoveCount int
}

func (s GameState) Finished() bool {
	return statemachine.IsTerminal(s.Status)
}
