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

	result, err := s.service.ImportTransactions(ctx, service.ImportTransactionsInput{
		UserID:  request.GetUserId(),
		CSVData: request.GetCsvData(),
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &ledgerv1.ImportTransactionsResponse{
		ImportedCount: int64(result.ImportedCount),
		FailedCount:   int64(result.FailedCount),
		Errors:        importTransactionErrors(result.Errors),
	}, nil
}

func importTransactionErrors(errors []service.ImportTransactionsError) []*ledgerv1.ImportTransactionError {
	result := make([]*ledgerv1.ImportTransactionError, 0, len(errors))
	for _, importError := range errors {
		result = append(result, &ledgerv1.ImportTransactionError{
			Row:     int64(importError.Row),
			Message: importError.Message,
		})
	}

	return result
}
