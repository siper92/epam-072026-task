package data

import (
	"context"
	"database/sql"
	"epam/task/game/data/gen"
	"errors"

	"epam/task/pkg/errs"
)

type sqlStore struct {
	q *gen.Queries
}

var _ Store = (*sqlStore)(nil)

func NewSQLStore(db gen.DBTX) Store {
	return &sqlStore{q: gen.New(db)}
}

func (s *sqlStore) CreateGame(ctx context.Context, game gen.Game) error {
	return s.q.CreateGame(ctx, gen.CreateGameParams(game))
}

func (s *sqlStore) GetGame(ctx context.Context, id string) (gen.Game, error) {
	game, err := s.q.GetGame(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return gen.Game{}, errs.Newf(errs.CodeNotFound, "game %q not found", id)
	}
	return game, err
}

func (s *sqlStore) UpdateGameState(ctx context.Context, id string, board string, status string) error {
	return s.q.UpdateGameState(ctx, gen.UpdateGameStateParams{Board: board, Status: status, ID: id})
}

func (s *sqlStore) SetPlayerO(ctx context.Context, id string, playerID string, status string) error {
	return s.q.SetPlayerO(ctx, gen.SetPlayerOParams{PlayerO: playerID, Status: status, ID: id})
}

func (s *sqlStore) CreatePlayer(ctx context.Context, player gen.Player) error {
	return s.q.CreatePlayer(ctx, gen.CreatePlayerParams(player))
}

func (s *sqlStore) GetPlayer(ctx context.Context, id string) (gen.Player, error) {
	player, err := s.q.GetPlayer(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return gen.Player{}, errs.Newf(errs.CodeNotFound, "player %q not found", id)
	}
	return player, err
}

func (s *sqlStore) ListWaitingGames(ctx context.Context, status string) ([]gen.Game, error) {
	return s.q.ListWaitingGames(ctx, status)
}

func (s *sqlStore) SaveToken(ctx context.Context, token string, expiresAt int64) error {
	return s.q.SaveToken(ctx, gen.SaveTokenParams{Token: token, ExpiresAt: expiresAt})
}

func (s *sqlStore) GetTokenExpiry(ctx context.Context, token string) (int64, error) {
	expiresAt, err := s.q.GetTokenExpiry(ctx, token)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errs.New(errs.CodeInvalidToken, "unknown token")
	}
	return expiresAt, err
}
