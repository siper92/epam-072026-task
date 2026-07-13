package db

import (
	"context"
)

const upsertGame = `
INSERT INTO games (id, player_x, player_o, grid, status, winner_id, move_count)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (id) DO UPDATE SET
    grid = excluded.grid,
    status = excluded.status,
    winner_id = excluded.winner_id,
    move_count = excluded.move_count
`

type UpsertGameParams struct {
	ID        string
	PlayerX   string
	PlayerO   string
	Grid      string
	Status    string
	WinnerID  string
	MoveCount int64
}

func (q *Queries) UpsertGame(ctx context.Context, arg UpsertGameParams) error {
	_, err := q.db.ExecContext(ctx, upsertGame,
		arg.ID,
		arg.PlayerX,
		arg.PlayerO,
		arg.Grid,
		arg.Status,
		arg.WinnerID,
		arg.MoveCount,
	)
	return err
}

const getGame = `
SELECT id, player_x, player_o, grid, status, winner_id, move_count
FROM games
WHERE id = ?
`

func (q *Queries) GetGame(ctx context.Context, id string) (Game, error) {
	row := q.db.QueryRowContext(ctx, getGame, id)
	var i Game
	err := row.Scan(
		&i.ID,
		&i.PlayerX,
		&i.PlayerO,
		&i.Grid,
		&i.Status,
		&i.WinnerID,
		&i.MoveCount,
	)
	return i, err
}

const listGamesByPlayer = `
SELECT id, player_x, player_o, grid, status, winner_id, move_count
FROM games
WHERE player_x = ? OR player_o = ?
ORDER BY id
`

type ListGamesByPlayerParams struct {
	PlayerX string
	PlayerO string
}

func (q *Queries) ListGamesByPlayer(ctx context.Context, arg ListGamesByPlayerParams) ([]Game, error) {
	rows, err := q.db.QueryContext(ctx, listGamesByPlayer, arg.PlayerX, arg.PlayerO)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Game
	for rows.Next() {
		var i Game
		if err := rows.Scan(
			&i.ID,
			&i.PlayerX,
			&i.PlayerO,
			&i.Grid,
			&i.Status,
			&i.WinnerID,
			&i.MoveCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
