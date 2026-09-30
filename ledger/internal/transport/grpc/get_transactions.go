package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (s *Server) GetTransactions(
	ctx context.Context,
	request *ledgerv1.GetTransactionsRequest,
) (*ledgerv1.GetTransactionsResponse, error) {
	if request == nil {
		return nil, invalidRequestError()
	}

	input := service.GetTransactionsInput{
		UserID:   request.GetUserId(),
		Category: request.GetCategory(),
	}

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

	transactions, err := s.service.GetTransactions(ctx, input)
	if err != nil {
		return nil, mapServiceError(err)
	}

	response := &ledgerv1.GetTransactionsResponse{
		Transactions: make([]*ledgerv1.Transaction, 0, len(transactions)),
	}
	for _, transaction := range transactions {
		response.Transactions = append(
			response.Transactions,
			transactionToProto(transaction),
		)
	}

	return response, nil
}

func transactionToProto(transaction domain.Transaction) *ledgerv1.Transaction {
	return &ledgerv1.Transaction{
		Id:          transaction.ID,
		UserId:      transaction.UserID,
		Amount:      transaction.Amount,
		Category:    transaction.Category,
		Description: transaction.Description,
		OccurredAt:  timestamppb.New(transaction.OccurredAt),
		CreatedAt:   timestamppb.New(transaction.CreatedAt),
	}
}
