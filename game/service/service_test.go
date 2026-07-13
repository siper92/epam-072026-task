package service

import (
	"context"
	"epam/task/game/data/gen"
	"os"
	"testing"

	"epam/task/game/auth"
	"epam/task/game/data"
	"epam/task/game/state_machine"
	"epam/task/pkg/config"
	"epam/task/pkg/errs"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-secret")
	config.LoadEnv()

	os.Exit(m.Run())
}

func newTestService() GameService {
	store := data.NewMemoryStore()
	return NewGameService(store, store, auth.NewService(store))
}

func TestCreatePublicGameListedInLobby(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	game, token, err := svc.CreateGame(ctx, "player-1", true)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if game.Status != string(state_machine.StatusWaitingForPlayers) {
		t.Fatalf("unexpected status %q", game.Status)
	}
	if game.Code != "" {
		t.Fatalf("public game should have no code, got %q", game.Code)
	}

	gameToken, err := svc.ValidateGameToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidateGameToken: %v", err)
	}
	if gameToken.GameID != game.ID || gameToken.PlayerID != "player-1" || gameToken.Mark != "X" {
		t.Fatalf("unexpected token claims: %+v", gameToken)
	}

	waiting, err := svc.WaitingGames(ctx)
	if err != nil {
		t.Fatalf("WaitingGames: %v", err)
	}
	if len(waiting) != 1 || waiting[0].ID != game.ID {
		t.Fatalf("expected game %q in lobby, got %v", game.ID, waiting)
	}
}

func TestCreateGameRequiresPlayer(t *testing.T) {
	svc := newTestService()

	if _, _, err := svc.CreateGame(context.Background(), "", true); !errs.HasCode(err, errs.CodeInvalidInput) {
		t.Fatalf("expected CodeInvalidInput, got %v", err)
	}
}

func TestPrivateGameHiddenAndNeedsCode(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	game, _, err := svc.CreateGame(ctx, "player-1", false)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if game.Code == "" {
		t.Fatal("private game should have a code")
	}

	waiting, err := svc.WaitingGames(ctx)
	if err != nil {
		t.Fatalf("WaitingGames: %v", err)
	}
	if len(waiting) != 0 {
		t.Fatalf("private game should not be listed, got %v", waiting)
	}

	if _, _, err = svc.JoinGame(ctx, game.ID, "player-2", "wrong-code"); !errs.HasCode(err, errs.CodeInvalidInput) {
		t.Fatalf("expected CodeInvalidInput for wrong code, got %v", err)
	}

	joined, _, err := svc.JoinGame(ctx, game.ID, "player-2", game.Code)
	if err != nil {
		t.Fatalf("JoinGame: %v", err)
	}
	if joined.PlayerO != "player-2" || joined.Status != string(state_machine.StatusPlayerXTurn) {
		t.Fatalf("unexpected game after join: %+v", joined)
	}
}

func TestJoinGameRules(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	game, _, err := svc.CreateGame(ctx, "player-1", true)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	if _, _, err = svc.JoinGame(ctx, game.ID, "player-1", ""); !errs.HasCode(err, errs.CodeInvalidInput) {
		t.Fatalf("expected CodeInvalidInput for creator joining, got %v", err)
	}
	if _, _, err = svc.JoinGame(ctx, "missing", "player-2", ""); !errs.HasCode(err, errs.CodeNotFound) {
		t.Fatalf("expected CodeNotFound, got %v", err)
	}

	joined, token, err := svc.JoinGame(ctx, game.ID, "player-2", "")
	if err != nil {
		t.Fatalf("JoinGame: %v", err)
	}

	gameToken, err := svc.ValidateGameToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidateGameToken: %v", err)
	}
	if gameToken.Mark != "O" || gameToken.GameID != joined.ID {
		t.Fatalf("unexpected token claims: %+v", gameToken)
	}

	if _, _, err = svc.JoinGame(ctx, game.ID, "player-3", ""); !errs.HasCode(err, errs.CodeInvalidTransition) {
		t.Fatalf("expected CodeInvalidTransition for full game, got %v", err)
	}
}

func TestMakeMoveFlow(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	game, tokenX, err := svc.CreateGame(ctx, "player-1", true)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	if _, err = svc.MakeMove(ctx, tokenX, 0, 0); !errs.HasCode(err, errs.CodeInvalidTransition) {
		t.Fatalf("expected CodeInvalidTransition before game start, got %v", err)
	}

	_, tokenO, err := svc.JoinGame(ctx, game.ID, "player-2", "")
	if err != nil {
		t.Fatalf("JoinGame: %v", err)
	}

	if _, err = svc.MakeMove(ctx, tokenO, 0, 0); !errs.HasCode(err, errs.CodeInvalidTransition) {
		t.Fatalf("expected CodeInvalidTransition when moving out of turn, got %v", err)
	}
	if _, err = svc.MakeMove(ctx, "not-a-token", 0, 0); !errs.HasCode(err, errs.CodeInvalidToken) {
		t.Fatalf("expected CodeInvalidToken, got %v", err)
	}

	updated, err := svc.MakeMove(ctx, tokenX, 0, 0)
	if err != nil {
		t.Fatalf("MakeMove: %v", err)
	}
	if updated.Board != "X________" || updated.Status != string(state_machine.StatusPlayerOTurn) {
		t.Fatalf("unexpected game after move: %+v", updated)
	}

	loaded, err := svc.GetGame(ctx, game.ID)
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if loaded.Board != updated.Board || loaded.Status != updated.Status {
		t.Fatalf("state was not persisted: %+v", loaded)
	}
}

func TestMakeMoveUntilWin(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	game, tokenX, err := svc.CreateGame(ctx, "player-1", true)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	_, tokenO, err := svc.JoinGame(ctx, game.ID, "player-2", "")
	if err != nil {
		t.Fatalf("JoinGame: %v", err)
	}

	moves := []struct {
		token    string
		row, col int
	}{
		{tokenX, 0, 0},
		{tokenO, 1, 0},
		{tokenX, 0, 1},
		{tokenO, 1, 1},
		{tokenX, 0, 2},
	}
	var last gen.Game
	for _, m := range moves {
		if last, err = svc.MakeMove(ctx, m.token, m.row, m.col); err != nil {
			t.Fatalf("MakeMove(%d,%d): %v", m.row, m.col, err)
		}
	}

	if last.Status != string(state_machine.StatusGameOverPlayerXWin) {
		t.Fatalf("expected X win, got %q", last.Status)
	}
	if _, err = svc.MakeMove(ctx, tokenO, 2, 2); !errs.HasCode(err, errs.CodeGameFinished) {
		t.Fatalf("expected CodeGameFinished, got %v", err)
	}
}
