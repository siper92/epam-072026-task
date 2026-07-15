package internal

import (
	"context"

	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/session"
)

type SessionAuth interface {
	Session() (session.Data, error)
	Login(ctx context.Context) (session.Data, error)
	Refresh(ctx context.Context) (session.Data, error)
}

type Lobby interface {
	WaitingGames(ctx context.Context) ([]api.GameResponse, error)
	CreateGame(ctx context.Context, public bool) (api.GameResponse, error)
	JoinGame(ctx context.Context, id string, code string) (api.GameResponse, error)
}

type GamePlay interface {
	GetGame(ctx context.Context, id string) (api.GameResponse, error)
	Move(ctx context.Context, id string, row int, col int) (api.GameResponse, error)
}

type GameClient interface {
	SessionAuth
	Lobby
	GamePlay
}

var _ GameClient = (*Client)(nil)
