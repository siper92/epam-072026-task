-- name: CreateUser :one
INSERT INTO users (id, username, password_hash, role)
VALUES (?, ?, ?, ?)
RETURNING id, username, password_hash, role, created_at;

-- name: GetUserByID :one
SELECT id, username, password_hash, role, created_at
FROM users
WHERE id = ?;

-- name: GetUserByUsername :one
SELECT id, username, password_hash, role, created_at
FROM users
WHERE username = ?;
