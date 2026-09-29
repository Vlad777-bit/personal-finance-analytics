package service

import (
	"context"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func (s *service) CreateBudget(
	ctx context.Context,
	input CreateBudgetInput,
) (domain.Budget, error) {
	budget, err := domain.NewBudget(domain.NewBudgetParams{
		UserID:   input.UserID,
		Category: input.Category,
		Limit:    input.Limit,
	})
	if err != nil {
		return domain.Budget{}, fmt.Errorf("validate budget: %w", err)
	}

	savedBudget, err := s.budgetRepository.Upsert(ctx, budget)
	if err != nil {
		return domain.Budget{}, fmt.Errorf("upsert budget: %w", err)
	}

	return savedBudget, nil
}
