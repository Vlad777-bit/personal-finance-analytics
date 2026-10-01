package transaction

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func (r *Repository) CreateWithinBudget(
	ctx context.Context,
	transaction domain.Transaction,
) (_ domain.Transaction, returnedErr error) {
	databaseTransaction, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("begin transaction creation: %w", err)
	}

	defer func() {
		rollbackErr := databaseTransaction.Rollback(ctx)
		if rollbackErr != nil && !errors.Is(rollbackErr, db.ErrTxClosed) {
			returnedErr = errors.Join(
				returnedErr,
				fmt.Errorf("rollback transaction creation: %w", rollbackErr),
			)
		}
	}()

	budget, budgetConfigured, err := lockBudget(
		ctx,
		databaseTransaction,
		transaction.UserID,
		transaction.Category,
	)
	if err != nil {
		return domain.Transaction{}, err
	}

	if budgetConfigured {
		from, to := monthBounds(transaction.OccurredAt)
		spent, sumErr := sumByCategoryAndPeriod(
			ctx,
			databaseTransaction,
			transaction.UserID,
			transaction.Category,
			from,
			to,
		)
		if sumErr != nil {
			return domain.Transaction{}, sumErr
		}

		if err := budget.ValidateSpending(spent, transaction.Amount); err != nil {
			return domain.Transaction{}, err
		}
	}

	createdTransaction, err := create(ctx, databaseTransaction, transaction)
	if err != nil {
		return domain.Transaction{}, err
	}

	if err := databaseTransaction.Commit(ctx); err != nil {
		return domain.Transaction{}, fmt.Errorf("commit transaction creation: %w", err)
	}

	return createdTransaction, nil
}

func lockBudget(
	ctx context.Context,
	databaseTransaction db.Tx,
	userID string,
	category string,
) (domain.Budget, bool, error) {
	const query = `
		SELECT limit_amount
		FROM budgets
		WHERE user_id = $1
		  AND category = $2
		FOR UPDATE
	`

	var limit int64
	err := databaseTransaction.QueryRow(ctx, query, userID, category).Scan(&limit)
	if errors.Is(err, db.ErrNoRows) {
		return domain.Budget{}, false, nil
	}
	if err != nil {
		return domain.Budget{}, false, fmt.Errorf("lock budget: %w", err)
	}

	return domain.Budget{
		UserID: userID, Category: category, Limit: limit,
	}, true, nil
}

func monthBounds(value time.Time) (time.Time, time.Time) {
	from := time.Date(
		value.Year(),
		value.Month(),
		1,
		0,
		0,
		0,
		0,
		value.Location(),
	)

	return from, from.AddDate(0, 1, 0)
}
