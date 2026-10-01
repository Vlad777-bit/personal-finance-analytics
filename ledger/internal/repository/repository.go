package repository

import (
	"context"
	"time"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

type TransactionRepository interface {
	CreateWithinBudget(
		ctx context.Context,
		transaction domain.Transaction,
	) (domain.Transaction, error)

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

	List(
		ctx context.Context,
		filter TransactionFilter,
	) ([]domain.Transaction, error)
}

type TransactionFilter struct {
	UserID   string
	Category string
	From     time.Time
	To       time.Time
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

	ListByUser(
		ctx context.Context,
		userID string,
	) ([]domain.Budget, error)
}

type ReportRepository interface {
	GetSummaryData(
		ctx context.Context,
		userID string,
		from time.Time,
		to time.Time,
	) ([]domain.CategoryReportData, error)
}
