package pgx

import (
	"context"
	"errors"
	"fmt"

	pgxdriver "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database"
)

const uniqueViolationCode = "23505"

var _ database.DB = (*Client)(nil)

type Client struct {
	pool *pgxpool.Pool
}

type row struct {
	row pgxdriver.Row
}

func New(ctx context.Context, dsn string) (*Client, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Client{pool: pool}, nil
}

func (c *Client) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) database.Row {
	return &row{row: c.pool.QueryRow(ctx, query, args...)}
}

func (c *Client) Close() {
	c.pool.Close()
}

func (r *row) Scan(dest ...any) error {
	if err := mapError(r.row.Scan(dest...)); err != nil {
		return fmt.Errorf("scan row: %w", err)
	}

	return nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgxdriver.ErrNoRows) {
		return database.ErrNoRows
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == uniqueViolationCode {
		return database.ErrUniqueViolation
	}

	return err
}
