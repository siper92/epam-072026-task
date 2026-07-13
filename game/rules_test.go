package game

import (
	"epam/task/game/state_machine"
	"testing"

	"epam/task/pkg/errs"
)

func gridFromRows(rows [GridSize]string) Grid {
	var grid Grid
	for r, row := range rows {
		for c, cell := range row {
			grid[r][c] = Mark(cell)
		}
	}
	return grid
}

func TestHasWinAllLines(t *testing.T) {
	cases := []struct {
		name   string
		winner Mark
		rows   [GridSize]string
	}{
		{"X row 0", MarkX, [GridSize]string{"XXX", "OO_", "___"}},
		{"X row 1", MarkX, [GridSize]string{"O_O", "XXX", "___"}},
		{"X row 2", MarkX, [GridSize]string{"_OO", "___", "XXX"}},
		{"X col 0", MarkX, [GridSize]string{"XO_", "XO_", "X__"}},
		{"X col 1", MarkX, [GridSize]string{"OX_", "_XO", "_X_"}},
		{"X col 2", MarkX, [GridSize]string{"O_X", "O_X", "__X"}},
		{"X diagonal", MarkX, [GridSize]string{"XO_", "OX_", "__X"}},
		{"X anti-diagonal", MarkX, [GridSize]string{"_OX", "OX_", "X__"}},
		{"O row 0", MarkO, [GridSize]string{"OOO", "XX_", "__X"}},
		{"O row 1", MarkO, [GridSize]string{"XX_", "OOO", "__X"}},
		{"O row 2", MarkO, [GridSize]string{"X_X", "_X_", "OOO"}},
		{"O col 0", MarkO, [GridSize]string{"OXX", "OX_", "O__"}},
		{"O col 1", MarkO, [GridSize]string{"XOX", "_O_", "XO_"}},
		{"O col 2", MarkO, [GridSize]string{"XXO", "_XO", "__O"}},
		{"O diagonal", MarkO, [GridSize]string{"OXX", "XO_", "__O"}},
		{"O anti-diagonal", MarkO, [GridSize]string{"XXO", "XO_", "O__"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			loser := MarkO
			if tc.winner == MarkO {
				loser = MarkX
			}
			grid := gridFromRows(tc.rows)
			if !hasWin(grid, tc.winner) {
				t.Errorf("expected win for %q on %v", tc.winner, tc.rows)
			}
			if hasWin(grid, loser) {
				t.Errorf("did not expect win for %q on %v", loser, tc.rows)
			}
		})
	}
}

func TestHasWinNoLine(t *testing.T) {
	cases := []struct {
		name string
		grid Grid
	}{
		{"empty grid", NewGrid()},
		{"single move", gridFromRows([GridSize]string{"X__", "___", "___"})},
		{"two in a row each", gridFromRows([GridSize]string{"XX_", "OO_", "___"})},
		{"blocked lines", gridFromRows([GridSize]string{"XOX", "OOX", "XXO"})},
		{"full draw", gridFromRows([GridSize]string{"OXO", "XOX", "XOX"})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if hasWin(tc.grid, MarkX) {
				t.Error("did not expect win for X")
			}
			if hasWin(tc.grid, MarkO) {
				t.Error("did not expect win for O")
			}
		})
	}
}

func TestIsFull(t *testing.T) {
	cases := []struct {
		name string
		grid Grid
		want bool
	}{
		{"full grid", gridFromRows([GridSize]string{"XOX", "OOX", "XXO"}), true},
		{"empty last cell", gridFromRows([GridSize]string{"XOX", "OOX", "XX_"}), false},
		{"empty first cell", gridFromRows([GridSize]string{"_OX", "OOX", "XXO"}), false},
		{"empty center", gridFromRows([GridSize]string{"XOX", "O_X", "XXO"}), false},
		{"empty grid", NewGrid(), false},
		{"single mark", gridFromRows([GridSize]string{"X__", "___", "___"}), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isFull(tc.grid); got != tc.want {
				t.Errorf("expected isFull=%v, got %v", tc.want, got)
			}
		})
	}
}

func TestPlayerMark(t *testing.T) {
	state := &GameState{ID: "g1", PlayerX: "alice", PlayerO: "bob"}
	if mark, err := playerMark(state, "alice"); err != nil || mark != MarkX {
		t.Errorf("expected MarkX for alice, got %q, %v", mark, err)
	}
	if mark, err := playerMark(state, "bob"); err != nil || mark != MarkO {
		t.Errorf("expected MarkO for bob, got %q, %v", mark, err)
	}
	for _, unknown := range []string{"mallory", "", "Alice", "BOB"} {
		if _, err := playerMark(state, unknown); !errs.HasCode(err, errs.CodeUnknownPlayer) {
			t.Errorf("expected code %q for %q, got %v", errs.CodeUnknownPlayer, unknown, err)
		}
	}
}

func TestValidateMove(t *testing.T) {
	grid := gridFromRows([GridSize]string{"X__", "_O_", "___"})
	cases := []struct {
		name   string
		status state_machine.State
		mark   Mark
		row    int
		col    int
		code   errs.Code
	}{
		{"valid O move", state_machine.StateTurnO, MarkO, 1, 0, ""},
		{"valid X move", state_machine.StateTurnX, MarkX, 2, 2, ""},
		{"X out of turn", state_machine.StateTurnO, MarkX, 2, 2, errs.CodeOutOfTurn},
		{"O out of turn", state_machine.StateTurnX, MarkO, 2, 2, errs.CodeOutOfTurn},
		{"cell occupied by X", state_machine.StateTurnO, MarkO, 0, 0, errs.CodeCellOccupied},
		{"cell occupied by O", state_machine.StateTurnX, MarkX, 1, 1, errs.CodeCellOccupied},
		{"row out of bounds", state_machine.StateTurnO, MarkO, GridSize, 0, errs.CodeOutOfBounds},
		{"negative row", state_machine.StateTurnO, MarkO, -1, 0, errs.CodeOutOfBounds},
		{"col out of bounds", state_machine.StateTurnO, MarkO, 0, GridSize, errs.CodeOutOfBounds},
		{"negative col", state_machine.StateTurnO, MarkO, 0, -1, errs.CodeOutOfBounds},
		{"both out of bounds", state_machine.StateTurnO, MarkO, -1, GridSize, errs.CodeOutOfBounds},
		{"far out of bounds", state_machine.StateTurnO, MarkO, 100, 100, errs.CodeOutOfBounds},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := &GameState{Grid: grid, Status: tc.status}
			err := validateMove(state, tc.mark, tc.row, tc.col)
			if errs.CodeOf(err) != tc.code {
				t.Errorf("expected code %q, got %v", tc.code, err)
			}
		})
	}
}

func TestApplyMove(t *testing.T) {
	state := &GameState{Grid: NewGrid()}
	applyMove(state, MarkX, 1, 2)
	if state.Grid[1][2] != MarkX {
		t.Errorf("expected MarkX at (1,2), got %q", state.Grid[1][2])
	}
	if state.MoveCount != 1 {
		t.Errorf("expected move count 1, got %d", state.MoveCount)
	}
	applyMove(state, MarkO, 0, 0)
	if state.Grid[0][0] != MarkO {
		t.Errorf("expected MarkO at (0,0), got %q", state.Grid[0][0])
	}
	if state.Grid[1][2] != MarkX {
		t.Errorf("expected MarkX to remain at (1,2), got %q", state.Grid[1][2])
	}
	if state.MoveCount != 2 {
		t.Errorf("expected move count 2, got %d", state.MoveCount)
	}
	applyMove(state, MarkX, 2, 1)
	if state.MoveCount != 3 {
		t.Errorf("expected move count 3, got %d", state.MoveCount)
	}
}

func TestMoveEvent(t *testing.T) {
	cases := []struct {
		name string
		rows [GridSize]string
		mark Mark
		want state_machine.Event
	}{
		{"X wins row", [GridSize]string{"XXX", "OO_", "___"}, MarkX, state_machine.EventWinX},
		{"X wins diagonal", [GridSize]string{"XO_", "OX_", "__X"}, MarkX, state_machine.EventWinX},
		{"O wins row", [GridSize]string{"XX_", "OOO", "X__"}, MarkO, state_machine.EventWinO},
		{"O wins column", [GridSize]string{"OXX", "OX_", "O__"}, MarkO, state_machine.EventWinO},
		{"draw on full grid", [GridSize]string{"XOX", "OOX", "XXO"}, MarkX, state_machine.EventDraw},
		{"X keeps playing", [GridSize]string{"X__", "___", "___"}, MarkX, state_machine.EventMoveX},
		{"O keeps playing", [GridSize]string{"XO_", "___", "___"}, MarkO, state_machine.EventMoveO},
		{"X wins on last cell", [GridSize]string{"XOO", "OXX", "XOX"}, MarkX, state_machine.EventWinX},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := moveEvent(gridFromRows(tc.rows), tc.mark); got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
