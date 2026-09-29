package domain

import "strings"

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
