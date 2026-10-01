package database

import (
	"context"
	"errors"
)

var (
	ErrNoRows   = errors.New("no rows")
	ErrTxClosed = errors.New("transaction is closed")
)

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

type QueryExecutor interface {
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
}

type Tx interface {
	QueryExecutor
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type DB interface {
	QueryExecutor
	Begin(ctx context.Context) (Tx, error)

	Close()
}
