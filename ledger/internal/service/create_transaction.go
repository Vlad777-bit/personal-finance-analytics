package service

import (
	"context"
	"errors"
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

	budget, err := s.budgetRepository.GetByCategory(
		ctx,
		transaction.UserID,
		transaction.Category,
	)
	if err != nil && !errors.Is(err, domain.ErrBudgetNotFound) {
		return domain.Transaction{}, fmt.Errorf("get budget: %w", err)
	}

	if err == nil {
		from, to := monthBounds(transaction.OccurredAt)

		spent, sumErr := s.transactionRepository.SumByCategoryAndPeriod(
			ctx,
			transaction.UserID,
			transaction.Category,
			from,
			to,
		)
		if sumErr != nil {
			return domain.Transaction{}, fmt.Errorf(
				"sum transactions by category and period: %w",
				sumErr,
			)
		}

		if spent+transaction.Amount > budget.Limit {
			return domain.Transaction{}, fmt.Errorf(
				"%w: category=%s limit=%d spent=%d amount=%d",
				domain.ErrBudgetExceeded,
				transaction.Category,
				budget.Limit,
				spent,
				transaction.Amount,
			)
		}
	}

	createdTransaction, err := s.transactionRepository.Create(ctx, transaction)
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
