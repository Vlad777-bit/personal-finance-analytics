package transaction

import (
	"context"
	"fmt"
	"time"

	db "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database"
)

func (r *Repository) SumByCategoryAndPeriod(
	ctx context.Context,
	userID string,
	category string,
	from time.Time,
	to time.Time,
) (int64, error) {
	return sumByCategoryAndPeriod(ctx, r.db, userID, category, from, to)
}

func sumByCategoryAndPeriod(
	ctx context.Context,
	database db.QueryExecutor,
	userID string,
	category string,
	from time.Time,
	to time.Time,
) (int64, error) {
	const query = `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1
		  AND category = $2
		  AND occurred_at >= $3
		  AND occurred_at < $4
	`

	var total int64

	err := database.QueryRow(
		ctx,
		query,
		userID,
		category,
		from,
		to,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf(
			"sum transactions by category and period: %w",
			err,
		)
	}

	return total, nil
}
