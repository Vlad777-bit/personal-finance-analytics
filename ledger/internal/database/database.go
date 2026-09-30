package database

import (
	"context"
	"errors"
)

var ErrNoRows = errors.New("no rows")

type Row interface {
	Scan(dest ...any) error
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}

type Result interface {
	RowsAffected() int64
}

type DB interface {
	Query(
		ctx context.Context,
		query string,
		args ...any,
	) (Rows, error)

	QueryRow(
		ctx context.Context,
		query string,
		args ...any,
	) Row

	Exec(
		ctx context.Context,
		query string,
		args ...any,
	) (Result, error)

	Close()
}
