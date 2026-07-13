package db

import (
	"context"
)

type Querier interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
	UpsertGame(ctx context.Context, arg UpsertGameParams) error
	GetGame(ctx context.Context, id string) (Game, error)
	ListGamesByPlayer(ctx context.Context, arg ListGamesByPlayerParams) ([]Game, error)
}

var _ Querier = (*Queries)(nil)
