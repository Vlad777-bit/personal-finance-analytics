package service

import (
	"context"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func (s *service) CreateTransaction(
	ctx context.Context,
	input CreateTransactionInput,
) (domain.Transaction, error) {
	transaction, err := domain.NewTransaction(domain.NewTransactionParams{
		UserID:      input.UserID,
		Amount:      input.Amount,
		Category:    input.Category,
		Description: input.Description,
		OccurredAt:  input.OccurredAt,
	})
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("validate transaction: %w", err)
	}

	createdTransaction, err := s.transactionRepository.CreateWithinBudget(ctx, transaction)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("create transaction: %w", err)
	}

	if err := s.summaryCache.InvalidateUser(ctx, createdTransaction.UserID); err != nil {
		s.logger.Warn(
			"invalidate summary cache after transaction",
			"user_id",
			createdTransaction.UserID,
			"error",
			err,
		)
	}

	return createdTransaction, nil
}
