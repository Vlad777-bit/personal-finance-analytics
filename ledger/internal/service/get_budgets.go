package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func (s *service) GetBudgets(
	ctx context.Context,
	input GetBudgetsInput,
) ([]domain.Budget, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return nil, domain.ErrUserIDRequired
	}

	budgets, err := s.budgetRepository.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}

	return budgets, nil
}
