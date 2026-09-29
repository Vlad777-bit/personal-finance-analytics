package repository

import (
	"context"
	"time"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

type TransactionRepository interface {
	Create(
		ctx context.Context,
		transaction domain.Transaction,
	) (domain.Transaction, error)

	SumByCategoryAndPeriod(
		ctx context.Context,
		userID string,
		category string,
		from time.Time,
		to time.Time,
	) (int64, error)
}

type BudgetRepository interface {
	GetByCategory(
		ctx context.Context,
		userID string,
		category string,
	) (domain.Budget, error)

	Upsert(
		ctx context.Context,
		budget domain.Budget,
	) (domain.Budget, error)
}
