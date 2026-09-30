package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

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
		Transaction: &ledgerv1.Transaction{
			Id:          transaction.ID,
			UserId:      transaction.UserID,
			Amount:      transaction.Amount,
			Category:    transaction.Category,
			Description: transaction.Description,
			OccurredAt:  timestamppb.New(transaction.OccurredAt),
			CreatedAt:   timestamppb.New(transaction.CreatedAt),
		},
	}, nil
}
