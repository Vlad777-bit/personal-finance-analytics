package service

import (
	"context"
	"time"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository"
)

var _ LedgerService = (*service)(nil)

type LedgerService interface {
	CreateTransaction(
		ctx context.Context,
		input CreateTransactionInput,
	) (domain.Transaction, error)
}

type CreateTransactionInput struct {
	UserID      string
	Amount      int64
	Category    string
	Description string
	OccurredAt  time.Time
}

type service struct {
	transactionRepository repository.TransactionRepository
	budgetRepository      repository.BudgetRepository
}

func New(
	transactionRepository repository.TransactionRepository,
	budgetRepository repository.BudgetRepository,
) LedgerService {
	return &service{
		transactionRepository: transactionRepository,
		budgetRepository:      budgetRepository,
	}
}

func monthBounds(value time.Time) (time.Time, time.Time) {
	from := time.Date(
		value.Year(),
		value.Month(),
		1,
		0,
		0,
		0,
		0,
		value.Location(),
	)

	return from, from.AddDate(0, 1, 0)
}
