-- name: DeactivatePlayerTokens :exec
UPDATE tokens SET active = FALSE WHERE player_id = ? AND active = TRUE;

-- name: SaveToken :exec
INSERT INTO tokens (token, player_id, expires_at, active) VALUES (?, ?, ?, TRUE);

-- name: GetTokenExpiry :one
SELECT expires_at FROM tokens WHERE token = ? AND active = TRUE;
