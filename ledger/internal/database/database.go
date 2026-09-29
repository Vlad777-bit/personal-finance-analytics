package database

import (
	"context"
	"errors"
)

var ErrNoRows = errors.New("no rows")

type Row interface {
	Scan(dest ...any) error
}

type DB interface {
	QueryRow(
		ctx context.Context,
		query string,
		args ...any,
	) Row

	Close()
}
