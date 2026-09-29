package domain

import (
	"strings"
	"time"
)

type Transaction struct {
	ID          string
	UserID      string
	Amount      int64
	Category    string
	Description string
	OccurredAt  time.Time
	CreatedAt   time.Time
}

type NewTransactionParams struct {
	UserID      string
	Amount      int64
	Category    string
	Description string
	OccurredAt  time.Time
}

func NewTransaction(params NewTransactionParams) (Transaction, error) {
	userID := strings.TrimSpace(params.UserID)
	if userID == "" {
		return Transaction{}, ErrUserIDRequired
	}

	if params.Amount <= 0 {
		return Transaction{}, ErrInvalidAmount
	}

	category := strings.TrimSpace(params.Category)
	if category == "" {
		return Transaction{}, ErrCategoryRequired
	}

	if params.OccurredAt.IsZero() {
		return Transaction{}, ErrDateRequired
	}

	return Transaction{
		UserID:      userID,
		Amount:      params.Amount,
		Category:    category,
		Description: strings.TrimSpace(params.Description),
		OccurredAt:  params.OccurredAt,
	}, nil
}
