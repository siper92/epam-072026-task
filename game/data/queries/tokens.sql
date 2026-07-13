-- name: SaveToken :exec
INSERT INTO tokens (token, player_id, expires_at) VALUES (?, ?, ?)
ON CONFLICT(player_id) DO UPDATE SET
    token = excluded.token,
    expires_at = excluded.expires_at;

-- name: GetTokenExpiry :one
SELECT expires_at FROM tokens WHERE token = ?;
