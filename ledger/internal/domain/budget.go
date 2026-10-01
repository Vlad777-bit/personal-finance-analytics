package domain

import (
	"fmt"
	"strings"
)

type Budget struct {
	ID       string
	UserID   string
	Category string
	Limit    int64
}

type NewBudgetParams struct {
	UserID   string
	Category string
	Limit    int64
}

func NewBudget(params NewBudgetParams) (Budget, error) {
	userID := strings.TrimSpace(params.UserID)
	if userID == "" {
		return Budget{}, ErrUserIDRequired
	}

	category := strings.TrimSpace(params.Category)
	if category == "" {
		return Budget{}, ErrCategoryRequired
	}

	if params.Limit <= 0 {
		return Budget{}, ErrInvalidBudget
	}

	return Budget{
		UserID:   userID,
		Category: category,
		Limit:    params.Limit,
	}, nil
}

func (b Budget) ValidateSpending(spent, amount int64) error {
	if amount > b.Limit || spent > b.Limit-amount {
		return fmt.Errorf(
			"%w: category=%s limit=%d spent=%d amount=%d",
			ErrBudgetExceeded,
			b.Category,
			b.Limit,
			spent,
			amount,
		)
	}

	return nil
}
