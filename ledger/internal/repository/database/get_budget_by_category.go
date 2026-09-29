package database

import (
	"context"
	"errors"
	"fmt"

	db "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func (r *Repository) GetByCategory(
	ctx context.Context,
	userID string,
	category string,
) (domain.Budget, error) {
	const query = `
		SELECT
			id,
			user_id,
			category,
			limit_amount
		FROM budgets
		WHERE user_id = $1
		  AND category = $2
	`

	var budget domain.Budget

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		category,
	).Scan(
		&budget.ID,
		&budget.UserID,
		&budget.Category,
		&budget.Limit,
	)
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			return domain.Budget{}, domain.ErrBudgetNotFound
		}

		return domain.Budget{}, fmt.Errorf(
			"query budget by category: %w",
			err,
		)
	}

	return budget, nil
}
