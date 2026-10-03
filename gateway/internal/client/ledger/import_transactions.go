package ledger

import (
	"context"

	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (c *Client) ImportTransactions(
	ctx context.Context,
	input ImportTransactionsInput,
) (ImportTransactionsResult, error) {
	result := ImportTransactionsResult{}
	response, err := c.service.ImportTransactions(
		ctx,
		&ledgerv1.ImportTransactionsRequest{
			UserId:  input.UserID,
			CsvData: input.CSVData,
		},
	)
	if err != nil {
		return result, mapError(err)
	}
	if response == nil {
		return result, ErrInvalidResponse
	}

	result.ImportedCount = int(response.GetImportedCount())
	result.FailedCount = int(response.GetFailedCount())
	for _, importError := range response.GetErrors() {
		result.Errors = append(result.Errors, ImportTransactionsError{
			Row:     importError.GetRow(),
			Message: importError.GetMessage(),
		})
	}

	return result, nil
}
