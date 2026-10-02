package grpc

import (
	"context"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (s *Server) ExportTransactions(
	ctx context.Context,
	request *ledgerv1.ExportTransactionsRequest,
) (*ledgerv1.ExportTransactionsResponse, error) {
	if request == nil {
		return nil, invalidRequestError()
	}

	input := service.ExportTransactionsInput{
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

	csvData, err := s.service.ExportTransactions(ctx, input)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &ledgerv1.ExportTransactionsResponse{CsvData: csvData}, nil
}
