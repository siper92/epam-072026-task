package data

import (
	"context"
	"database/sql"
	_ "embed"

	"ticTacSolved/task/pkg/errs"
)

//go:embed schema/schema.sql
var schemaSQL string

func OpenSQLStore(
	ctx context.Context,
	driver string,
	dsn string,
) (Store, *sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, nil, errs.Wrap(
			errs.CodeInvalidInput,
			"failed to open database",
			err,
		)
	}
	if err = ensureSchema(ctx, db); err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	return NewSQLStore(db), db, nil
}

func ensureSchema(ctx context.Context, db *sql.DB) error {
	row := db.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'games'",
	)
	var count int
	if err := row.Scan(&count); err != nil {
		return errs.Wrap(
			errs.CodeInvalidInput,
			"failed to inspect database schema",
			err,
		)
	}
	if count > 0 {
		return nil
	}
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return errs.Wrap(
			errs.CodeInvalidInput,
			"failed to apply database schema",
			err,
		)
	}
	return nil
}
