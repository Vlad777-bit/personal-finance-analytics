package pgx

import (
	"context"
	"errors"
	"fmt"

	pgxdriver "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database"
)

var _ database.DB = (*Client)(nil)

type Client struct {
	pool *pgxpool.Pool
}

type row struct {
	row pgxdriver.Row
}

type rows struct {
	rows pgxdriver.Rows
}

func New(
	ctx context.Context,
	dsn string,
) (*Client, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Client{
		pool: pool,
	}, nil
}

func (c *Client) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) database.Row {
	return &row{
		row: c.pool.QueryRow(ctx, query, args...),
	}
}

func (c *Client) Query(
	ctx context.Context,
	query string,
	args ...any,
) (database.Rows, error) {
	result, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query rows: %w", err)
	}

	return &rows{rows: result}, nil
}

func (c *Client) Close() {
	c.pool.Close()
}

func (r *row) Scan(dest ...any) error {
	err := r.row.Scan(dest...)
	if errors.Is(err, pgxdriver.ErrNoRows) {
		return database.ErrNoRows
	}

	if err != nil {
		return fmt.Errorf("scan row: %w", err)
	}

	return nil
}

func (r *rows) Next() bool {
	return r.rows.Next()
}

func (r *rows) Scan(dest ...any) error {
	if err := r.rows.Scan(dest...); err != nil {
		return fmt.Errorf("scan row: %w", err)
	}

	return nil
}

func (r *rows) Err() error {
	if err := r.rows.Err(); err != nil {
		return fmt.Errorf("iterate rows: %w", err)
	}

	return nil
}

func (r *rows) Close() {
	r.rows.Close()
}
