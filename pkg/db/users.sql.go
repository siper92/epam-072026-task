package db

import (
	"context"
)

const createUser = `
INSERT INTO users (id, username, password_hash, role)
VALUES (?, ?, ?, ?)
RETURNING id, username, password_hash, role, created_at
`

type CreateUserParams struct {
	ID           string
	Username     string
	PasswordHash string
	Role         string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.ID,
		arg.Username,
		arg.PasswordHash,
		arg.Role,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.PasswordHash,
		&i.Role,
		&i.CreatedAt,
	)
	return i, err
}

const getUserByID = `
SELECT id, username, password_hash, role, created_at
FROM users
WHERE id = ?
`

func (q *Queries) GetUserByID(ctx context.Context, id string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.PasswordHash,
		&i.Role,
		&i.CreatedAt,
	)
	return i, err
}

const getUserByUsername = `
SELECT id, username, password_hash, role, created_at
FROM users
WHERE username = ?
`

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByUsername, username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.PasswordHash,
		&i.Role,
		&i.CreatedAt,
	)
	return i, err
}
