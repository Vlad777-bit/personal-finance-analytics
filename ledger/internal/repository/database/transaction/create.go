package transaction

import (
	"context"
	"fmt"

	db "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func (r *Repository) Create(
	ctx context.Context,
	transaction domain.Transaction,
) (domain.Transaction, error) {
	return create(ctx, r.db, transaction)
}

func create(
	ctx context.Context,
	database db.QueryExecutor,
	transaction domain.Transaction,
) (domain.Transaction, error) {
	const query = `
		INSERT INTO transactions (
			user_id,
			amount,
			category,
			description,
			occurred_at
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			user_id,
			amount,
			category,
			description,
			occurred_at,
			created_at
	`

	var savedTransaction domain.Transaction

	err := database.QueryRow(
		ctx,
		query,
		transaction.UserID,
		transaction.Amount,
		transaction.Category,
		transaction.Description,
		transaction.OccurredAt,
	).Scan(
		&savedTransaction.ID,
		&savedTransaction.UserID,
		&savedTransaction.Amount,
		&savedTransaction.Category,
		&savedTransaction.Description,
		&savedTransaction.OccurredAt,
		&savedTransaction.CreatedAt,
	)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf(
			"create transaction: %w",
			err,
		)
	}

	return savedTransaction, nil
}
