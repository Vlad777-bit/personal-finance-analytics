package ledger

import (
	"errors"
	"time"

	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

type CreateTransactionInput struct {
	UserID      string
	Amount      int64
	Category    string
	Description string
	OccurredAt  time.Time
}

type Transaction struct {
	ID          string
	UserID      string
	Amount      int64
	Category    string
	Description string
	OccurredAt  time.Time
	CreatedAt   time.Time
}

type GetTransactionsInput struct {
	UserID   string
	Category string
	From     time.Time
	To       time.Time
}

type CreateBudgetInput struct {
	UserID   string
	Category string
	Limit    int64
}

type Budget struct {
	ID       string
	UserID   string
	Category string
	Limit    int64
}

type GetBudgetsInput struct {
	UserID string
}

func transactionFromProto(transaction *ledgerv1.Transaction) (Transaction, error) {
	if transaction == nil {
		return Transaction{}, ErrInvalidResponse
	}
	if transaction.GetOccurredAt() == nil || transaction.GetCreatedAt() == nil {
		return Transaction{}, errors.Join(
			ErrInvalidResponse,
			errors.New("transaction timestamps are required"),
		)
	}

	return Transaction{
		ID:          transaction.GetId(),
		UserID:      transaction.GetUserId(),
		Amount:      transaction.GetAmount(),
		Category:    transaction.GetCategory(),
		Description: transaction.GetDescription(),
		OccurredAt:  transaction.GetOccurredAt().AsTime(),
		CreatedAt:   transaction.GetCreatedAt().AsTime(),
	}, nil
}

func budgetFromProto(budget *ledgerv1.Budget) (Budget, error) {
	if budget == nil {
		return Budget{}, ErrInvalidResponse
	}

	return Budget{
		ID:       budget.GetId(),
		UserID:   budget.GetUserId(),
		Category: budget.GetCategory(),
		Limit:    budget.GetLimitAmount(),
	}, nil
}
