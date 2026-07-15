package handlers

import (
	"context"

	"ticTacSolved/task/game/service"
	"ticTacSolved/task/pkg/api"
)

type LoginResult struct {
	PlayerID string
	Session  api.Token
	Refresh  api.Token
}

type Authenticator interface {
	Login(
		ctx context.Context,
		user string,
		password string,
		sessionTTL int64,
		refreshTTL int64,
	) (LoginResult, error)
	Refresh(
		ctx context.Context,
		refreshToken string,
		sessionTTL int64,
	) (api.Token, error)
}

type Handlers struct {
	games service.GameService
	auth  Authenticator
	queue service.QueueService
}

func New(
	games service.GameService,
	auth Authenticator,
	queue service.QueueService,
) *Handlers {
	return &Handlers{games: games, auth: auth, queue: queue}
}
