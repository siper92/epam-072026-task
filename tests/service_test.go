package tests

import (
	"testing"

	"epam/task/game"
	"epam/task/pkg/errs"
)

func newTestService(t *testing.T) (game.GameService, game.GameState) {
	t.Helper()
	svc := game.NewService(game.NewMemoryStore())
	state, err := svc.NewGame("alice", "bob")
	if err != nil {
		t.Fatalf("NewGame failed: %v", err)
	}
	return svc, state
}

func move(t *testing.T, svc game.GameService, gameID, playerID string, row, col int) game.GameState {
	t.Helper()
	state, err := svc.GameAction(gameID, game.Action{PlayerID: playerID, Type: game.ActionMove, Row: row, Col: col})
	if err != nil {
		t.Fatalf("move by %s at (%d,%d) failed: %v", playerID, row, col, err)
	}
	return state
}

func TestNewGame(t *testing.T) {
	_, state := newTestService(t)
	if state.ID == "" {
		t.Error("expected non-empty game ID")
	}
	if state.Status != state_machine.StateTurnX {
		t.Errorf("expected initial status %q, got %q", state_machine.StateTurnX, state.Status)
	}
	if state.PlayerX != "alice" || state.PlayerO != "bob" {
		t.Errorf("unexpected players: %q, %q", state.PlayerX, state.PlayerO)
	}
	if state.MoveCount != 0 {
		t.Errorf("expected 0 moves, got %d", state.MoveCount)
	}
}

func TestNewGameInvalidInput(t *testing.T) {
	svc := game.NewService(game.NewMemoryStore())
	cases := []struct {
		name    string
		playerX string
		playerO string
	}{
		{"empty player X", "", "bob"},
		{"empty player O", "alice", ""},
		{"both empty", "", ""},
		{"same players", "alice", "alice"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.NewGame(tc.playerX, tc.playerO)
			if !errs.HasCode(err, errs.CodeInvalidInput) {
				t.Errorf("expected code %q, got error %v", errs.CodeInvalidInput, err)
			}
		})
	}
}

func TestXWinsRow(t *testing.T) {
	svc, initial := newTestService(t)
	move(t, svc, initial.ID, "alice", 0, 0)
	move(t, svc, initial.ID, "bob", 1, 0)
	move(t, svc, initial.ID, "alice", 0, 1)
	move(t, svc, initial.ID, "bob", 1, 1)
	final := move(t, svc, initial.ID, "alice", 0, 2)

	if final.Status != state_machine.StateWonX {
		t.Errorf("expected status %q, got %q", state_machine.StateWonX, final.Status)
	}
	if final.WinnerID != "alice" {
		t.Errorf("expected winner alice, got %q", final.WinnerID)
	}
}

func TestOWinsColumn(t *testing.T) {
	svc, initial := newTestService(t)
	move(t, svc, initial.ID, "alice", 0, 0)
	move(t, svc, initial.ID, "bob", 0, 2)
	move(t, svc, initial.ID, "alice", 1, 0)
	move(t, svc, initial.ID, "bob", 1, 2)
	move(t, svc, initial.ID, "alice", 2, 1)
	final := move(t, svc, initial.ID, "bob", 2, 2)

	if final.Status != state_machine.StateWonO {
		t.Errorf("expected status %q, got %q", state_machine.StateWonO, final.Status)
	}
	if final.WinnerID != "bob" {
		t.Errorf("expected winner bob, got %q", final.WinnerID)
	}
}

func TestXWinsDiagonal(t *testing.T) {
	svc, initial := newTestService(t)
	move(t, svc, initial.ID, "alice", 1, 1)
	move(t, svc, initial.ID, "bob", 0, 1)
	move(t, svc, initial.ID, "alice", 0, 0)
	move(t, svc, initial.ID, "bob", 0, 2)
	final := move(t, svc, initial.ID, "alice", 2, 2)

	if final.Status != state_machine.StateWonX {
		t.Errorf("expected status %q, got %q", state_machine.StateWonX, final.Status)
	}
	if final.WinnerID != "alice" {
		t.Errorf("expected winner alice, got %q", final.WinnerID)
	}
	if final.MoveCount != 5 {
		t.Errorf("expected 5 moves, got %d", final.MoveCount)
	}
}

func TestOWinsAntiDiagonal(t *testing.T) {
	svc, initial := newTestService(t)
	move(t, svc, initial.ID, "alice", 0, 0)
	move(t, svc, initial.ID, "bob", 1, 1)
	move(t, svc, initial.ID, "alice", 0, 1)
	move(t, svc, initial.ID, "bob", 0, 2)
	move(t, svc, initial.ID, "alice", 2, 2)
	final := move(t, svc, initial.ID, "bob", 2, 0)

	if final.Status != state_machine.StateWonO {
		t.Errorf("expected status %q, got %q", state_machine.StateWonO, final.Status)
	}
	if final.WinnerID != "bob" {
		t.Errorf("expected winner bob, got %q", final.WinnerID)
	}
	if final.MoveCount != 6 {
		t.Errorf("expected 6 moves, got %d", final.MoveCount)
	}
}

func TestDraw(t *testing.T) {
	svc, initial := newTestService(t)
	move(t, svc, initial.ID, "alice", 0, 0)
	move(t, svc, initial.ID, "bob", 0, 1)
	move(t, svc, initial.ID, "alice", 0, 2)
	move(t, svc, initial.ID, "bob", 1, 1)
	move(t, svc, initial.ID, "alice", 1, 0)
	move(t, svc, initial.ID, "bob", 2, 0)
	move(t, svc, initial.ID, "alice", 1, 2)
	move(t, svc, initial.ID, "bob", 2, 2)
	final := move(t, svc, initial.ID, "alice", 2, 1)

	if final.Status != state_machine.StateDraw {
		t.Errorf("expected status %q, got %q", state_machine.StateDraw, final.Status)
	}
	if final.WinnerID != "" {
		t.Errorf("expected no winner, got %q", final.WinnerID)
	}
	if final.MoveCount != 9 {
		t.Errorf("expected 9 moves, got %d", final.MoveCount)
	}
}

func TestForfeit(t *testing.T) {
	cases := []struct {
		name       string
		forfeiter  string
		wantStatus state_machine.State
		wantWinner string
	}{
		{"X forfeits", "alice", state_machine.StateWonO, "bob"},
		{"O forfeits", "bob", state_machine.StateWonX, "alice"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, initial := newTestService(t)
			final, err := svc.GameAction(initial.ID, game.Action{PlayerID: tc.forfeiter, Type: game.ActionForfeit})
			if err != nil {
				t.Fatalf("forfeit failed: %v", err)
			}
			if final.Status != tc.wantStatus {
				t.Errorf("expected status %q, got %q", tc.wantStatus, final.Status)
			}
			if final.WinnerID != tc.wantWinner {
				t.Errorf("expected winner %q, got %q", tc.wantWinner, final.WinnerID)
			}
		})
	}
}

func TestForfeitOutOfTurn(t *testing.T) {
	svc, initial := newTestService(t)
	move(t, svc, initial.ID, "alice", 0, 0)
	final, err := svc.GameAction(initial.ID, game.Action{PlayerID: "alice", Type: game.ActionForfeit})
	if err != nil {
		t.Fatalf("forfeit on opponent's turn failed: %v", err)
	}
	if final.Status != state_machine.StateWonO {
		t.Errorf("expected status %q, got %q", state_machine.StateWonO, final.Status)
	}
	if final.WinnerID != "bob" {
		t.Errorf("expected winner bob, got %q", final.WinnerID)
	}
}

func TestInvalidMoves(t *testing.T) {
	svc, initial := newTestService(t)
	move(t, svc, initial.ID, "alice", 1, 1)

	cases := []struct {
		name   string
		action game.Action
		code   errs.Code
	}{
		{"occupied cell", game.Action{PlayerID: "bob", Type: game.ActionMove, Row: 1, Col: 1}, errs.CodeCellOccupied},
		{"out of turn", game.Action{PlayerID: "alice", Type: game.ActionMove, Row: 0, Col: 0}, errs.CodeOutOfTurn},
		{"row too big", game.Action{PlayerID: "bob", Type: game.ActionMove, Row: 3, Col: 0}, errs.CodeOutOfBounds},
		{"negative row", game.Action{PlayerID: "bob", Type: game.ActionMove, Row: -1, Col: 0}, errs.CodeOutOfBounds},
		{"col too big", game.Action{PlayerID: "bob", Type: game.ActionMove, Row: 0, Col: 3}, errs.CodeOutOfBounds},
		{"negative col", game.Action{PlayerID: "bob", Type: game.ActionMove, Row: 0, Col: -1}, errs.CodeOutOfBounds},
		{"far out of bounds", game.Action{PlayerID: "bob", Type: game.ActionMove, Row: 100, Col: 100}, errs.CodeOutOfBounds},
		{"unknown player", game.Action{PlayerID: "mallory", Type: game.ActionMove, Row: 0, Col: 0}, errs.CodeUnknownPlayer},
		{"empty player", game.Action{PlayerID: "", Type: game.ActionMove, Row: 0, Col: 0}, errs.CodeUnknownPlayer},
		{"unknown player forfeit", game.Action{PlayerID: "mallory", Type: game.ActionForfeit}, errs.CodeUnknownPlayer},
		{"unknown action", game.Action{PlayerID: "bob", Type: "DANCE"}, errs.CodeInvalidAction},
		{"empty action", game.Action{PlayerID: "bob", Type: ""}, errs.CodeInvalidAction},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.GameAction(initial.ID, tc.action)
			if !errs.HasCode(err, tc.code) {
				t.Errorf("expected code %q, got error %v", tc.code, err)
			}
		})
	}
}

func TestActionOnFinishedGame(t *testing.T) {
	finishByForfeit := func(t *testing.T, svc game.GameService, id string) {
		t.Helper()
		if _, err := svc.GameAction(id, game.Action{PlayerID: "bob", Type: game.ActionForfeit}); err != nil {
			t.Fatalf("forfeit failed: %v", err)
		}
	}
	finishByWin := func(t *testing.T, svc game.GameService, id string) {
		t.Helper()
		move(t, svc, id, "alice", 0, 0)
		move(t, svc, id, "bob", 1, 0)
		move(t, svc, id, "alice", 0, 1)
		move(t, svc, id, "bob", 1, 1)
		move(t, svc, id, "alice", 0, 2)
	}
	cases := []struct {
		name   string
		finish func(*testing.T, game.GameService, string)
		action game.Action
	}{
		{"move after forfeit", finishByForfeit, game.Action{PlayerID: "alice", Type: game.ActionMove, Row: 0, Col: 0}},
		{"forfeit after forfeit", finishByForfeit, game.Action{PlayerID: "alice", Type: game.ActionForfeit}},
		{"move after win", finishByWin, game.Action{PlayerID: "bob", Type: game.ActionMove, Row: 2, Col: 2}},
		{"forfeit after win", finishByWin, game.Action{PlayerID: "bob", Type: game.ActionForfeit}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, initial := newTestService(t)
			tc.finish(t, svc, initial.ID)
			_, err := svc.GameAction(initial.ID, tc.action)
			if !errs.HasCode(err, errs.CodeGameFinished) {
				t.Errorf("expected code %q, got error %v", errs.CodeGameFinished, err)
			}
		})
	}
}

func TestGameNotFound(t *testing.T) {
	svc, _ := newTestService(t)
	for _, id := range []string{"missing", "sdasd"} {
		_, err := svc.GetState(id)
		if !errs.HasCode(err, errs.CodeGameNotFound) {
			t.Errorf("expected code %q for GetState(%q), got error %v", errs.CodeGameNotFound, id, err)
		}
		_, err = svc.GameAction(id, game.Action{PlayerID: "alice", Type: game.ActionMove})
		if !errs.HasCode(err, errs.CodeGameNotFound) {
			t.Errorf("expected code %q for GameAction(%q), got error %v", errs.CodeGameNotFound, id, err)
		}
	}
}

func TestGetStateRoundtrip(t *testing.T) {
	svc, initial := newTestService(t)
	afterMove := move(t, svc, initial.ID, "alice", 2, 2)

	loaded, err := svc.GetState(initial.ID)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if loaded != afterMove {
		t.Errorf("GetState mismatch:\n got  %+v\n want %+v", loaded, afterMove)
	}
}

func TestReturnedStateIsCopy(t *testing.T) {
	svc, initial := newTestService(t)
	state := move(t, svc, initial.ID, "alice", 0, 0)
	state.Grid[2][2] = 'O'
	state.WinnerID = "mallory"

	loaded, err := svc.GetState(initial.ID)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if loaded.Grid[2][2] != game.MarkEmpty || loaded.WinnerID != "" {
		t.Error("mutating a returned state must not affect stored state")
	}
}

func TestMultipleSimultaneousGames(t *testing.T) {
	svc := game.NewService(game.NewMemoryStore())
	first, err := svc.NewGame("alice", "bob")
	if err != nil {
		t.Fatalf("NewGame failed: %v", err)
	}
	second, err := svc.NewGame("carol", "dave")
	if err != nil {
		t.Fatalf("NewGame failed: %v", err)
	}
	if first.ID == second.ID {
		t.Fatal("expected distinct game IDs")
	}
	move(t, svc, first.ID, "alice", 0, 0)
	loadedSecond, err := svc.GetState(second.ID)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if loadedSecond.MoveCount != 0 {
		t.Error("moves in one game must not affect another game")
	}
	move(t, svc, second.ID, "carol", 1, 1)
	loadedFirst, err := svc.GetState(first.ID)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if loadedFirst.MoveCount != 1 {
		t.Errorf("expected 1 move in first game, got %d", loadedFirst.MoveCount)
	}
	if loadedFirst.Grid[1][1] != game.MarkEmpty {
		t.Error("a move in one game must not mark another game's grid")
	}
}
