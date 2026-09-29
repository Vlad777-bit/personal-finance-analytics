package service

import (
	"context"
	"time"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository"
)

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
) *service {
	return &service{
		transactionRepository: transactionRepository,
		budgetRepository:      budgetRepository,
	}
}
