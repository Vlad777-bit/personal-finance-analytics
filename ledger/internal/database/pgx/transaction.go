package pgx

import (
	"context"
	"errors"
	"fmt"

	pgxdriver "github.com/jackc/pgx/v5"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database"
)

var _ database.Tx = (*transaction)(nil)

type transaction struct {
	tx pgxdriver.Tx
}

func (c *Client) Begin(ctx context.Context) (database.Tx, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	return &transaction{tx: tx}, nil
}

func (t *transaction) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) database.Row {
	return &row{row: t.tx.QueryRow(ctx, query, args...)}
}

func (t *transaction) Query(
	ctx context.Context,
	query string,
	args ...any,
) (database.Rows, error) {
	result, err := t.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query transaction rows: %w", err)
	}

	return &rows{rows: result}, nil
}

func (t *transaction) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (database.Result, error) {
	commandTag, err := t.tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("execute transaction query: %w", err)
	}

	return &result{rowsAffected: commandTag.RowsAffected()}, nil
}

func (t *transaction) Commit(ctx context.Context) error {
	if err := t.tx.Commit(ctx); err != nil {
		if errors.Is(err, pgxdriver.ErrTxClosed) {
			return database.ErrTxClosed
		}

		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (t *transaction) Rollback(ctx context.Context) error {
	if err := t.tx.Rollback(ctx); err != nil {
		if errors.Is(err, pgxdriver.ErrTxClosed) {
			return database.ErrTxClosed
		}

		return fmt.Errorf("rollback transaction: %w", err)
	}

	return nil
}
