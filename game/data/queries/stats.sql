-- name: AddStats :exec
INSERT INTO stats (player_id, wins, losses, draws)
VALUES (?, ?, ?, ?)
ON CONFLICT (player_id) DO UPDATE SET
    wins = wins + excluded.wins,
    losses = losses + excluded.losses,
    draws = draws + excluded.draws;

-- name: ListLeaders :many
SELECT player_id, wins, losses, draws FROM stats
ORDER BY wins DESC, draws DESC, losses ASC, player_id ASC
LIMIT ?;
