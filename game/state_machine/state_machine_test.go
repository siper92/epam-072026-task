package state_machine

import (
	"testing"

	"ticTacSolved/task/pkg/errs"
)

func TestParseGameState(t *testing.T) {
	cases := []struct {
		name          string
		encoded       string
		wantState     GameStatus
		wantPlayer    string
		wantMoveCount int
	}{
		{"empty board", "_________", StatusPlayerXTurn, "X", 0},
		{"mid-game O turn", "X___O___X", StatusPlayerOTurn, "O", 3},
		{"mid-game X turn", "XO_______", StatusPlayerXTurn, "X", 2},
		{"X wins top row", "XXXOO____", StatusGameOverPlayerXWin, "", 5},
		{"O wins column", "OXXOX_O__", StatusGameOverPlayerOWin, "", 6},
		{"O wins middle row", "X_XOOO_X_", StatusGameOverPlayerOWin, "", 6},
		{"X wins main diagonal", "XOO_X___X", StatusGameOverPlayerXWin, "", 5},
		{"X wins anti-diagonal", "OOX_X_X__", StatusGameOverPlayerXWin, "", 5},
		{"draw", "XOXOOXXXO", StatusGameOverDraw, "", 9},
		{"lowercase input", "xo_______", StatusPlayerXTurn, "X", 2},
		{"surrounding whitespace", " \n\tX___O___X\t\n ", StatusPlayerOTurn, "O", 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state, err := parseGameState(tc.encoded)
			if err != nil {
				t.Fatalf("ParseGameState(%q) failed: %v", tc.encoded, err)
			}
			if state.State != tc.wantState {
				t.Errorf("expected state %q, got %q", tc.wantState, state.State)
			}
			if state.CurrentPlayer != tc.wantPlayer {
				t.Errorf("expected current player %q, got %q", tc.wantPlayer, state.CurrentPlayer)
			}
			if state.MoveCount != tc.wantMoveCount {
				t.Errorf("expected move count %d, got %d", tc.wantMoveCount, state.MoveCount)
			}
		})
	}
}

func TestParseGameStateInvalid(t *testing.T) {
	cases := []struct {
		name    string
		encoded string
	}{
		{"too short", "____"},
		{"too long", "__________"},
		{"bad character", "____Z____"},
		{"too many X", "XX_______"},
		{"too many O", "O________"},
		{"X wins with equal counts", "XXXOOO___"},
		{"both win", "XXXOOOXO_"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseGameState(tc.encoded)
			if !errs.HasCode(err, errs.CodeInvalidInput) {
				t.Errorf("expected code %q, got %v", errs.CodeInvalidInput, err)
			}
		})
	}
}

func TestProcessMoveInvalid(t *testing.T) {
	state, err := parseGameState("X________")
	if err != nil {
		t.Fatalf("ParseGameState failed: %v", err)
	}
	m := newStateMachine(state)
	if err := m.ProcessMove(0, 0); !errs.HasCode(err, errs.CodeCellOccupied) {
		t.Errorf("expected code %q, got %v", errs.CodeCellOccupied, err)
	}
	if err := m.ProcessMove(3, 0); !errs.HasCode(err, errs.CodeOutOfBounds) {
		t.Errorf("expected code %q, got %v", errs.CodeOutOfBounds, err)
	}
	if err := m.ProcessMove(0, -1); !errs.HasCode(err, errs.CodeOutOfBounds) {
		t.Errorf("expected code %q, got %v", errs.CodeOutOfBounds, err)
	}
	waiting := newStateMachine(GameState{State: StatusWaitingForPlayers})
	if err := waiting.ProcessMove(0, 0); !errs.HasCode(err, errs.CodeInvalidTransition) {
		t.Errorf("expected code %q, got %v", errs.CodeInvalidTransition, err)
	}
}

func TestProcessMoveFullGame(t *testing.T) {
	cases := []struct {
		name      string
		moves     [][2]int
		wantState GameStatus
	}{
		{
			"X wins top row",
			[][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}, {0, 2}},
			StatusGameOverPlayerXWin,
		},
		{
			"O wins middle row",
			[][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}, {2, 2}, {1, 2}},
			StatusGameOverPlayerOWin,
		},
		{
			"draw",
			[][2]int{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 2}, {1, 1}, {2, 0}, {2, 2}, {2, 1}},
			StatusGameOverDraw,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := NewStateMachine("_________")
			if err != nil {
				t.Fatalf("NewStateMachine failed: %v", err)
			}
			for i, mv := range tc.moves {
				if err := m.ProcessMove(mv[0], mv[1]); err != nil {
					t.Fatalf("move %d at (%d,%d) failed: %v", i, mv[0], mv[1], err)
				}
				state := m.GetCurrentState()
				if state.MoveCount != i+1 {
					t.Errorf("after move %d expected move count %d, got %d", i, i+1, state.MoveCount)
				}
				if i < len(tc.moves)-1 {
					wantState, wantPlayer := StatusPlayerOTurn, "O"
					if i%2 == 1 {
						wantState, wantPlayer = StatusPlayerXTurn, "X"
					}
					if state.State != wantState {
						t.Errorf("after move %d expected state %q, got %q", i, wantState, state.State)
					}
					if state.CurrentPlayer != wantPlayer {
						t.Errorf("after move %d expected current player %q, got %q", i, wantPlayer, state.CurrentPlayer)
					}
				}
			}
			final := m.GetCurrentState()
			if final.State != tc.wantState {
				t.Errorf("expected final state %q, got %q", tc.wantState, final.State)
			}
			if final.CurrentPlayer != "" {
				t.Errorf("expected no current player, got %q", final.CurrentPlayer)
			}
			if err := m.ProcessMove(2, 2); !errs.HasCode(err, errs.CodeGameFinished) {
				t.Errorf("expected code %q, got %v", errs.CodeGameFinished, err)
			}
		})
	}
}

func TestPlayerLeft(t *testing.T) {
	cases := []struct {
		name      string
		player    string
		wantState GameStatus
	}{
		{"X leaves", "X", StatusGameOverPlayerOWin},
		{"O leaves", "O", StatusGameOverPlayerXWin},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := NewStateMachine("X___O____")
			if err != nil {
				t.Fatalf("NewStateMachine failed: %v", err)
			}
			if err := m.PlayerLeft(tc.player); err != nil {
				t.Fatalf("PlayerLeft(%q) failed: %v", tc.player, err)
			}
			state := m.GetCurrentState()
			if state.State != tc.wantState {
				t.Errorf("expected state %q, got %q", tc.wantState, state.State)
			}
			if state.CurrentPlayer != "" {
				t.Errorf("expected no current player, got %q", state.CurrentPlayer)
			}
		})
	}
	t.Run("unknown player", func(t *testing.T) {
		m, err := NewStateMachine("_________")
		if err != nil {
			t.Fatalf("NewStateMachine failed: %v", err)
		}
		if err := m.PlayerLeft("Z"); !errs.HasCode(err, errs.CodeInvalidInput) {
			t.Errorf("expected code %q, got %v", errs.CodeInvalidInput, err)
		}
	})
	t.Run("after game over", func(t *testing.T) {
		m, err := NewStateMachine("XXXOO____")
		if err != nil {
			t.Fatalf("NewStateMachine failed: %v", err)
		}
		if err := m.PlayerLeft("O"); !errs.HasCode(err, errs.CodeGameFinished) {
			t.Errorf("expected code %q, got %v", errs.CodeGameFinished, err)
		}
	})
}

func TestGetCurrentStateReturnsCopy(t *testing.T) {
	m, err := NewStateMachine("_________")
	if err != nil {
		t.Fatalf("NewStateMachine failed: %v", err)
	}
	state := m.GetCurrentState()
	state.Board[0][0] = "X"
	state.MoveCount = 5
	state.State = StatusGameOverDraw
	current := m.GetCurrentState()
	if current.Board[0][0] != "_" {
		t.Errorf("expected board cell (0,0) to be %q, got %q", "_", current.Board[0][0])
	}
	if current.MoveCount != 0 {
		t.Errorf("expected move count 0, got %d", current.MoveCount)
	}
	if current.State != StatusPlayerXTurn {
		t.Errorf("expected state %q, got %q", StatusPlayerXTurn, current.State)
	}
}
