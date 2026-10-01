package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (s *Server) GetSummary(
	ctx context.Context,
	request *ledgerv1.GetSummaryRequest,
) (*ledgerv1.GetSummaryResponse, error) {
	if request == nil {
		return nil, invalidRequestError()
	}

	input := service.GetSummaryInput{UserID: request.GetUserId()}
	if from := request.GetFrom(); from != nil {
		if err := from.CheckValid(); err != nil {
			return nil, invalidTimestampError("from", err)
		}

		input.From = from.AsTime()
	}
	if to := request.GetTo(); to != nil {
		if err := to.CheckValid(); err != nil {
			return nil, invalidTimestampError("to", err)
		}

		input.To = to.AsTime()
	}

	summary, err := s.service.GetSummary(ctx, input)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &ledgerv1.GetSummaryResponse{Summary: summaryToProto(summary)}, nil
}

func summaryToProto(summary domain.Summary) *ledgerv1.Summary {
	categories := make([]*ledgerv1.CategorySummary, 0, len(summary.Categories))
	for _, category := range summary.Categories {
		categories = append(categories, &ledgerv1.CategorySummary{
			Category:         category.Category,
			Spent:            category.Spent,
			BudgetLimit:      category.BudgetLimit,
			BudgetConfigured: category.BudgetConfigured,
			Remaining:        category.Remaining,
			BudgetExceeded:   category.BudgetExceeded,
		})
	}

	return &ledgerv1.Summary{
		UserId:     summary.UserID,
		From:       timestamppb.New(summary.From),
		To:         timestamppb.New(summary.To),
		TotalSpent: summary.TotalSpent,
		Categories: categories,
	}
}
