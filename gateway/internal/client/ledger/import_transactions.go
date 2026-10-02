package ledger

import (
	"context"

	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (c *Client) ImportTransactions(
	ctx context.Context,
	input ImportTransactionsInput,
) (int, error) {
	response, err := c.service.ImportTransactions(
		ctx,
		&ledgerv1.ImportTransactionsRequest{
			UserId:  input.UserID,
			CsvData: input.CSVData,
		},
	)
	if err != nil {
		return 0, mapError(err)
	}
	if response == nil {
		return 0, ErrInvalidResponse
	}

	return int(response.GetImportedCount()), nil
}
