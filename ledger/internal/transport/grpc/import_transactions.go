package grpc

import (
	"context"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (s *Server) ImportTransactions(
	ctx context.Context,
	request *ledgerv1.ImportTransactionsRequest,
) (*ledgerv1.ImportTransactionsResponse, error) {
	if request == nil {
		return nil, invalidRequestError()
	}

	importedCount, err := s.service.ImportTransactions(ctx, service.ImportTransactionsInput{
		UserID:  request.GetUserId(),
		CSVData: request.GetCsvData(),
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &ledgerv1.ImportTransactionsResponse{
		ImportedCount: int64(importedCount),
	}, nil
}
