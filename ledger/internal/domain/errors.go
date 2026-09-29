package domain

import "errors"

var (
	ErrUserIDRequired   = errors.New("user id is required")
	ErrCategoryRequired = errors.New("category is required")
	ErrInvalidAmount    = errors.New("amount must be greater than zero")
	ErrInvalidBudget    = errors.New("budget limit must be greater than zero")
	ErrDateRequired     = errors.New("transaction date is required")

	ErrBudgetNotFound = errors.New("budget not found")
	ErrBudgetExceeded = errors.New("budget exceeded")
)
