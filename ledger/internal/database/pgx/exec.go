package pgx

import (
	"context"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database"
)

type result struct {
	rowsAffected int64
}

func (r *result) RowsAffected() int64 {
	return r.rowsAffected
}

func (c *Client) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (database.Result, error) {
	commandTag, err := c.pool.Exec(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}

	return &result{
		rowsAffected: commandTag.RowsAffected(),
	}, nil
}
