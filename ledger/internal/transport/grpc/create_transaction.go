package grpc

import (
	"context"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (s *Server) CreateTransaction(
	ctx context.Context,
	request *ledgerv1.CreateTransactionRequest,
) (*ledgerv1.CreateTransactionResponse, error) {
	if request == nil {
		return nil, invalidRequestError()
	}

	input := service.CreateTransactionInput{
		UserID:      request.GetUserId(),
		Amount:      request.GetAmount(),
		Category:    request.GetCategory(),
		Description: request.GetDescription(),
	}
	if occurredAt := request.GetOccurredAt(); occurredAt != nil {
		if err := occurredAt.CheckValid(); err != nil {
			return nil, invalidTimestampError("occurred_at", err)
		}

		input.OccurredAt = occurredAt.AsTime()
	}

	transaction, err := s.service.CreateTransaction(ctx, input)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &ledgerv1.CreateTransactionResponse{
		Transaction: transactionToProto(transaction),
	}, nil
}
