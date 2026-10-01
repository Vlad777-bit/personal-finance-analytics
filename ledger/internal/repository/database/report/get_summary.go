package report

import (
	"context"
	"fmt"
	"time"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func (r *Repository) GetSummaryData(
	ctx context.Context,
	userID string,
	from time.Time,
	to time.Time,
) ([]domain.CategoryReportData, error) {
	rows, err := r.db.Query(
		ctx,
		`
			WITH transaction_totals AS (
				SELECT category, SUM(amount)::BIGINT AS spent
				FROM transactions
				WHERE user_id = $1
				  AND occurred_at >= $2
				  AND occurred_at < $3
				GROUP BY category
			),
			user_budgets AS (
				SELECT category, limit_amount
				FROM budgets
				WHERE user_id = $1
			)
			SELECT
				COALESCE(t.category, b.category) AS category,
				COALESCE(t.spent, 0)::BIGINT AS spent,
				COALESCE(b.limit_amount, 0)::BIGINT AS budget_limit,
				(b.category IS NOT NULL) AS budget_configured
			FROM transaction_totals t
			FULL OUTER JOIN user_budgets b ON b.category = t.category
			ORDER BY category ASC
		`,
		userID,
		from,
		to,
	)
	if err != nil {
		return nil, fmt.Errorf("query summary data: %w", err)
	}
	defer rows.Close()

	data := make([]domain.CategoryReportData, 0)
	for rows.Next() {
		var category domain.CategoryReportData
		if err = rows.Scan(
			&category.Category,
			&category.Spent,
			&category.BudgetLimit,
			&category.BudgetConfigured,
		); err != nil {
			return nil, fmt.Errorf("scan summary data: %w", err)
		}

		data = append(data, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate summary data: %w", err)
	}

	return data, nil
}
