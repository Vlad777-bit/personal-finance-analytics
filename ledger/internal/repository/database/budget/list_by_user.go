package budget

import (
	"context"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func (r *Repository) ListByUser(
	ctx context.Context,
	userID string,
) ([]domain.Budget, error) {
	rows, err := r.db.Query(
		ctx,
		`
			SELECT id, user_id, category, limit_amount
			FROM budgets
			WHERE user_id = $1
			ORDER BY category ASC, id ASC
		`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query budgets by user: %w", err)
	}
	defer rows.Close()

	budgets := make([]domain.Budget, 0)
	for rows.Next() {
		var budget domain.Budget
		if err := rows.Scan(
			&budget.ID,
			&budget.UserID,
			&budget.Category,
			&budget.Limit,
		); err != nil {
			return nil, fmt.Errorf("scan budget: %w", err)
		}

		budgets = append(budgets, budget)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate budgets: %w", err)
	}

	return budgets, nil
}
