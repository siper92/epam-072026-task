package state_machine

import (
	"testing"

	"epam/task/pkg/errs"
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
		{"draw", "XOXOOXXXO", StatusGameOverDraw, "", 9},
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
	waiting := newStateMachine(GameState{State: StatusWaitingForPlayers})
	if err := waiting.ProcessMove(0, 0); !errs.HasCode(err, errs.CodeInvalidTransition) {
		t.Errorf("expected code %q, got %v", errs.CodeInvalidTransition, err)
	}
}
