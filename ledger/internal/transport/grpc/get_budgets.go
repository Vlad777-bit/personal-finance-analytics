package grpc

import (
	"context"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (s *Server) GetBudgets(
	ctx context.Context,
	request *ledgerv1.GetBudgetsRequest,
) (*ledgerv1.GetBudgetsResponse, error) {
	if request == nil {
		return nil, invalidRequestError()
	}

	budgets, err := s.service.GetBudgets(
		ctx,
		service.GetBudgetsInput{UserID: request.GetUserId()},
	)
	if err != nil {
		return nil, mapServiceError(err)
	}

	response := &ledgerv1.GetBudgetsResponse{
		Budgets: make([]*ledgerv1.Budget, 0, len(budgets)),
	}
	for _, budget := range budgets {
		response.Budgets = append(response.Budgets, budgetToProto(budget))
	}

	return response, nil
}

func budgetToProto(budget domain.Budget) *ledgerv1.Budget {
	return &ledgerv1.Budget{
		Id:          budget.ID,
		UserId:      budget.UserID,
		Category:    budget.Category,
		LimitAmount: budget.Limit,
	}
}
