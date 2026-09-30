package grpc

import (
	"context"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (s *Server) CreateBudget(
	ctx context.Context,
	request *ledgerv1.CreateBudgetRequest,
) (*ledgerv1.CreateBudgetResponse, error) {
	if request == nil {
		return nil, invalidRequestError()
	}

	budget, err := s.service.CreateBudget(
		ctx,
		service.CreateBudgetInput{
			UserID:   request.GetUserId(),
			Category: request.GetCategory(),
			Limit:    request.GetLimitAmount(),
		},
	)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &ledgerv1.CreateBudgetResponse{
		Budget: &ledgerv1.Budget{
			Id:          budget.ID,
			UserId:      budget.UserID,
			Category:    budget.Category,
			LimitAmount: budget.Limit,
		},
	}, nil
}
