package game

import (
	"database/sql"
	"epam/task/game/state_machine"
	"errors"

	"epam/task/pkg/errs"
)

const sqliteSchema = `
CREATE TABLE IF NOT EXISTS games (
	id TEXT PRIMARY KEY,
	player_x TEXT NOT NULL,
	player_o TEXT NOT NULL,
	grid TEXT NOT NULL,
	status TEXT NOT NULL,
	winner_id TEXT NOT NULL DEFAULT '',
	move_count INTEGER NOT NULL DEFAULT 0
);`

const sqliteUpsertGame = `
INSERT INTO games (id, player_x, player_o, grid, status, winner_id, move_count)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	grid = excluded.grid,
	status = excluded.status,
	winner_id = excluded.winner_id,
	move_count = excluded.move_count;`

const sqliteSelectGame = `
SELECT id, player_x, player_o, grid, status, winner_id, move_count
FROM games WHERE id = ?;`

type SQLiteStore struct {
	db *sql.DB
}

var _ Store = (*SQLiteStore)(nil)

func NewSQLiteStore(db *sql.DB) (*SQLiteStore, error) {
	if db == nil {
		return nil, errs.New(errs.CodeInvalidInput, "database handle is required")
	}
	if _, err := db.Exec(sqliteSchema); err != nil {
		return nil, errs.Wrap(errs.CodeStorageFailure, "failed to initialize games table", err)
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Save(state GameState) error {
	if state.ID == "" {
		return errs.New(errs.CodeInvalidInput, "game state must have an ID")
	}
	_, err := s.db.Exec(sqliteUpsertGame,
		state.ID,
		state.PlayerX,
		state.PlayerO,
		state.Grid.Encode(),
		string(state.Status),
		state.WinnerID,
		state.MoveCount,
	)
	if err != nil {
		return errs.Wrap(errs.CodeStorageFailure, "failed to save game", err)
	}
	return nil
}

func (s *SQLiteStore) Load(gameID string) (GameState, error) {
	var state GameState
	var encodedGrid, status string
	err := s.db.QueryRow(sqliteSelectGame, gameID).Scan(
		&state.ID,
		&state.PlayerX,
		&state.PlayerO,
		&encodedGrid,
		&status,
		&state.WinnerID,
		&state.MoveCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return GameState{}, errs.Newf(errs.CodeGameNotFound, "game %q not found", gameID)
	}
	if err != nil {
		return GameState{}, errs.Wrap(errs.CodeStorageFailure, "failed to load game", err)
	}
	grid, err := ParseGrid(encodedGrid)
	if err != nil {
		return GameState{}, errs.Wrap(errs.CodeStorageFailure, "stored grid is corrupt", err)
	}
	state.Grid = grid
	state.Status = state_machine.State(status)
	if !state_machine.IsValid(state.Status) {
		return GameState{}, errs.Newf(errs.CodeStorageFailure, "stored status %q is corrupt", status)
	}
	return state, nil
}
