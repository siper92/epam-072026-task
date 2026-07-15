package service

import (
	"context"
	"testing"

	"ticTacSolved/task/game/data/gen"
	"ticTacSolved/task/pkg/errs"
)

type move struct {
	row, col int
}

func playGame(
	t *testing.T,
	svc GameService,
	playerX string,
	playerO string,
	moves []move,
) {
	t.Helper()
	ctx := context.Background()

	game, tokenX, err := svc.CreateGame(ctx, playerX, true)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	_, tokenO, err := svc.JoinGame(ctx, game.ID, playerO, "")
	if err != nil {
		t.Fatalf("JoinGame: %v", err)
	}

	token := tokenX
	for i, m := range moves {
		if _, err = svc.MakeMove(ctx, token, m.row, m.col); err != nil {
			t.Fatalf("MakeMove %d (%d,%d): %v", i, m.row, m.col, err)
		}
		if token == tokenX {
			token = tokenO
		} else {
			token = tokenX
		}
	}
}

var xWinMoves = []move{
	{0, 0}, {1, 0}, {0, 1}, {1, 1}, {0, 2},
}

var drawMoves = []move{
	{0, 0}, {0, 1}, {0, 2}, {1, 1}, {1, 0},
	{1, 2}, {2, 1}, {2, 0}, {2, 2},
}

func TestLeadersRecordsWins(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	playGame(t, svc, "winner", "loser", xWinMoves)

	leaders, err := svc.Leaders(ctx, 10)
	if err != nil {
		t.Fatalf("Leaders: %v", err)
	}
	want := []gen.Stat{
		{PlayerID: "winner", Wins: 1},
		{PlayerID: "loser", Losses: 1},
	}
	if len(leaders) != len(want) {
		t.Fatalf("leaders = %+v, want %+v", leaders, want)
	}
	for i, entry := range want {
		if leaders[i] != entry {
			t.Fatalf("leaders[%d] = %+v, want %+v", i, leaders[i], entry)
		}
	}
}

func TestLeadersRecordsDraws(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	playGame(t, svc, "player-1", "player-2", drawMoves)

	leaders, err := svc.Leaders(ctx, 10)
	if err != nil {
		t.Fatalf("Leaders: %v", err)
	}
	if len(leaders) != 2 {
		t.Fatalf("expected 2 entries, got %+v", leaders)
	}
	for _, leader := range leaders {
		if leader.Draws != 1 || leader.Wins != 0 || leader.Losses != 0 {
			t.Fatalf("unexpected stat %+v", leader)
		}
	}
}

func TestLeadersLimitAndOrder(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	playGame(t, svc, "top", "other", xWinMoves)
	playGame(t, svc, "top", "other", xWinMoves)

	leaders, err := svc.Leaders(ctx, 1)
	if err != nil {
		t.Fatalf("Leaders: %v", err)
	}
	if len(leaders) != 1 || leaders[0].PlayerID != "top" || leaders[0].Wins != 2 {
		t.Fatalf("unexpected leaders %+v", leaders)
	}

	if _, err = svc.Leaders(ctx, 0); err != nil {
		t.Fatalf("Leaders with default limit: %v", err)
	}
}

func TestUnfinishedGameRecordsNothing(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	game, tokenX, err := svc.CreateGame(ctx, "player-1", true)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, _, err = svc.JoinGame(ctx, game.ID, "player-2", ""); err != nil {
		t.Fatalf("JoinGame: %v", err)
	}
	if _, err = svc.MakeMove(ctx, tokenX, 0, 0); err != nil {
		t.Fatalf("MakeMove: %v", err)
	}

	leaders, err := svc.Leaders(ctx, 10)
	if err != nil {
		t.Fatalf("Leaders: %v", err)
	}
	if len(leaders) != 0 {
		t.Fatalf("expected no entries, got %+v", leaders)
	}

	if _, err = svc.MakeMove(ctx, "bad-token", 0, 0); !errs.HasCode(err, errs.CodeInvalidToken) {
		t.Fatalf("expected CodeInvalidToken, got %v", err)
	}
}
