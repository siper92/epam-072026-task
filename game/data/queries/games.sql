-- name: CreateGame :exec
INSERT INTO games (id, code, is_public, board, status, player_x, player_o)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetGame :one
SELECT * FROM games WHERE id = ?;

-- name: UpdateGameState :exec
UPDATE games SET board = ?, status = ? WHERE id = ?;

-- name: SetPlayerO :exec
UPDATE games SET player_o = ?, status = ? WHERE id = ?;

-- name: ListWaitingGames :many
SELECT * FROM games WHERE status = ? AND is_public = TRUE;
