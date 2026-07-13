-- name: CreatePlayer :exec
INSERT INTO players (id, name) VALUES (?, ?);

-- name: GetPlayer :one
SELECT * FROM players WHERE id = ?;
