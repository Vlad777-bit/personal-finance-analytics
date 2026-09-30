package budget

import (
	"context"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func (r *Repository) Upsert(
	ctx context.Context,
	budget domain.Budget,
) (domain.Budget, error) {
	const query = `
		INSERT INTO budgets (
			user_id,
			category,
			limit_amount
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, category)
		DO UPDATE SET
			limit_amount = EXCLUDED.limit_amount,
			updated_at = NOW()
		RETURNING
			id,
			user_id,
			category,
			limit_amount
	`

	var savedBudget domain.Budget

	err := r.db.QueryRow(
		ctx,
		query,
		budget.UserID,
		budget.Category,
		budget.Limit,
	).Scan(
		&savedBudget.ID,
		&savedBudget.UserID,
		&savedBudget.Category,
		&savedBudget.Limit,
	)
	if err != nil {
		return domain.Budget{}, fmt.Errorf(
			"upsert budget: %w",
			err,
		)
	}

	return savedBudget, nil
}
