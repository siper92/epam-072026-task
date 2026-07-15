package data

import (
	"context"
	"ticTacSolved/task/game/data/gen"
)

type GameStore interface {
	CreateGame(ctx context.Context, game gen.Game) error
	GetGame(ctx context.Context, id string) (gen.Game, error)
	UpdateGameState(ctx context.Context, id string, board string, status string) error
	SetPlayerO(ctx context.Context, id string, playerID string, status string) error
}

type PlayerStore interface {
	CreatePlayer(ctx context.Context, player gen.Player) error
	GetPlayer(ctx context.Context, id string) (gen.Player, error)
}

type LobbyStore interface {
	ListWaitingGames(ctx context.Context, status string) ([]gen.Game, error)
}

type TokenStore interface {
	SaveToken(ctx context.Context, playerID string, token string, expiresAt int64) error
	GetTokenExpiry(ctx context.Context, token string) (int64, error)
}

type StatsStore interface {
	RecordResult(ctx context.Context, winnerID string, loserID string, draw bool) error
	ListLeaders(ctx context.Context, limit int64) ([]gen.Stat, error)
}

type Store interface {
	GameStore
	PlayerStore
	LobbyStore
	TokenStore
	StatsStore
}
