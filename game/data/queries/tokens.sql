-- name: SaveToken :exec
INSERT INTO tokens (token, expires_at) VALUES (?, ?);

-- name: GetTokenExpiry :one
SELECT expires_at FROM tokens WHERE token = ?;
