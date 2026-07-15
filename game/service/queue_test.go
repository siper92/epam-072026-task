package service

import (
	"context"
	"testing"

	"ticTacSolved/task/game/state_machine"
	"ticTacSolved/task/pkg/errs"
)

func TestQueueJoin(t *testing.T) {
	t.Run("requires a player id", func(t *testing.T) {
		q := NewQueueService(newTestService())

		if _, _, err := q.Join(context.Background(), ""); !errs.HasCode(err, errs.CodeInvalidInput) {
			t.Fatalf("expected CodeInvalidInput, got %v", err)
		}
	})

	t.Run("first player creates a public waiting game", func(t *testing.T) {
		svc := newTestService()
		q := NewQueueService(svc)
		ctx := context.Background()

		game, token, err := q.Join(ctx, "player-1")
		if err != nil {
			t.Fatalf("Join: %v", err)
		}
		if !game.IsPublic || game.PlayerX != "player-1" {
			t.Fatalf("unexpected game: %+v", game)
		}
		if game.Status != string(state_machine.StatusWaitingForPlayers) {
			t.Fatalf("unexpected status %q", game.Status)
		}
		if token == "" {
			t.Fatal("expected a game token")
		}
	})

	t.Run("second player is paired with the waiting game", func(t *testing.T) {
		svc := newTestService()
		q := NewQueueService(svc)
		ctx := context.Background()

		created, _, err := q.Join(ctx, "player-1")
		if err != nil {
			t.Fatalf("Join player-1: %v", err)
		}
		joined, token, err := q.Join(ctx, "player-2")
		if err != nil {
			t.Fatalf("Join player-2: %v", err)
		}
		if joined.ID != created.ID {
			t.Fatalf("expected pairing with %q, got %q", created.ID, joined.ID)
		}
		if joined.PlayerO != "player-2" {
			t.Fatalf("unexpected player O %q", joined.PlayerO)
		}
		if joined.Status != string(state_machine.StatusPlayerXTurn) {
			t.Fatalf("unexpected status %q", joined.Status)
		}

		claims, err := svc.ValidateGameToken(ctx, token)
		if err != nil {
			t.Fatalf("ValidateGameToken: %v", err)
		}
		if claims.Mark != "O" || claims.PlayerID != "player-2" {
			t.Fatalf("unexpected token claims: %+v", claims)
		}
	})

	t.Run("a player is never paired with themselves", func(t *testing.T) {
		q := NewQueueService(newTestService())
		ctx := context.Background()

		first, _, err := q.Join(ctx, "player-1")
		if err != nil {
			t.Fatalf("first Join: %v", err)
		}
		second, _, err := q.Join(ctx, "player-1")
		if err != nil {
			t.Fatalf("second Join: %v", err)
		}
		if first.ID == second.ID {
			t.Fatal("player was paired with their own game")
		}
		if second.Status != string(state_machine.StatusWaitingForPlayers) {
			t.Fatalf("unexpected status %q", second.Status)
		}
	})
}
