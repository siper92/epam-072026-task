-- name: UpsertGame :exec
INSERT INTO games (id, player_x, player_o, grid, status, winner_id, move_count)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (id) DO UPDATE SET
    grid = excluded.grid,
    status = excluded.status,
    winner_id = excluded.winner_id,
    move_count = excluded.move_count;

-- name: GetGame :one
SELECT id, player_x, player_o, grid, status, winner_id, move_count
FROM games
WHERE id = ?;

-- name: ListGamesByPlayer :many
SELECT id, player_x, player_o, grid, status, winner_id, move_count
FROM games
WHERE player_x = ? OR player_o = ?
ORDER BY id;
