package database

import (
	"context"
	"fmt"
	"time"
)

func (r *Repository) SumByCategoryAndPeriod(
	ctx context.Context,
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

	err := r.db.QueryRow(
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
