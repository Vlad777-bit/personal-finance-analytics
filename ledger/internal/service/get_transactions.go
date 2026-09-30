package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository"
)

func (s *service) GetTransactions(
	ctx context.Context,
	input GetTransactionsInput,
) ([]domain.Transaction, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return nil, domain.ErrUserIDRequired
	}

	if input.From.IsZero() || input.To.IsZero() || !input.From.Before(input.To) {
		return nil, domain.ErrInvalidPeriod
	}

	transactions, err := s.transactionRepository.List(
		ctx,
		repository.TransactionFilter{
			UserID:   userID,
			Category: strings.TrimSpace(input.Category),
			From:     input.From,
			To:       input.To,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}

	return transactions, nil
}
