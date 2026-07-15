package service

import (
	"context"

	"ticTacSolved/task/game/data/gen"
	"ticTacSolved/task/pkg/errs"
)

type QueueService interface {
	Join(ctx context.Context, playerID string) (gen.Game, string, error)
}

type queueService struct {
	lobby Lobby
}

var _ QueueService = (*queueService)(nil)

func NewQueueService(lobby Lobby) QueueService {
	return &queueService{lobby: lobby}
}

func (q *queueService) Join(
	ctx context.Context,
	playerID string,
) (gen.Game, string, error) {
	if playerID == "" {
		return gen.Game{}, "", errs.New(errs.CodeInvalidInput, "player id is required")
	}

	waiting, err := q.lobby.WaitingGames(ctx)
	if err != nil {
		return gen.Game{}, "", err
	}
	for _, game := range waiting {
		if game.PlayerX == playerID {
			continue
		}

		return q.lobby.JoinGame(ctx, game.ID, playerID, "")
	}

	return q.lobby.CreateGame(ctx, playerID, true)
}
