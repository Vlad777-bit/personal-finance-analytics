package database

import (
	"context"
	"errors"
)

var (
	ErrNoRows          = errors.New("no rows")
	ErrUniqueViolation = errors.New("unique constraint violation")
)

type Row interface {
	Scan(dest ...any) error
}

type Result interface {
	RowsAffected() int64
}

type DB interface {
	QueryRow(ctx context.Context, query string, args ...any) Row
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	Close()
}
