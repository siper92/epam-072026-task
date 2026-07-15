package gen

import (
	"context"
)

type Stat struct {
	PlayerID string
	Wins     int64
	Losses   int64
	Draws    int64
}

const addStats = `-- name: AddStats :exec
INSERT INTO stats (player_id, wins, losses, draws)
VALUES (?, ?, ?, ?)
ON CONFLICT (player_id) DO UPDATE SET
    wins = wins + excluded.wins,
    losses = losses + excluded.losses,
    draws = draws + excluded.draws
`

type AddStatsParams struct {
	PlayerID string
	Wins     int64
	Losses   int64
	Draws    int64
}

func (q *Queries) AddStats(ctx context.Context, arg AddStatsParams) error {
	_, err := q.db.ExecContext(ctx, addStats,
		arg.PlayerID,
		arg.Wins,
		arg.Losses,
		arg.Draws,
	)
	return err
}

const listLeaders = `-- name: ListLeaders :many
SELECT player_id, wins, losses, draws FROM stats
ORDER BY wins DESC, draws DESC, losses ASC, player_id ASC
LIMIT ?
`

func (q *Queries) ListLeaders(ctx context.Context, limit int64) ([]Stat, error) {
	rows, err := q.db.QueryContext(ctx, listLeaders, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Stat
	for rows.Next() {
		var i Stat
		if err := rows.Scan(
			&i.PlayerID,
			&i.Wins,
			&i.Losses,
			&i.Draws,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
