package transaction

import (
	"context"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository"
)

func (r *Repository) List(
	ctx context.Context,
	filter repository.TransactionFilter,
) ([]domain.Transaction, error) {
	rows, err := r.db.Query(
		ctx,
		`
			SELECT id, user_id, amount, category, description,
			       occurred_at, created_at
			FROM transactions
			WHERE user_id = $1
			  AND occurred_at >= $2
			  AND occurred_at < $3
			  AND ($4 = '' OR category = $4)
			ORDER BY occurred_at DESC, created_at DESC, id DESC
		`,
		filter.UserID,
		filter.From,
		filter.To,
		filter.Category,
	)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]domain.Transaction, 0)
	for rows.Next() {
		var transaction domain.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.Category,
			&transaction.Description,
			&transaction.OccurredAt,
			&transaction.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}

		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions: %w", err)
	}

	return transactions, nil
}
