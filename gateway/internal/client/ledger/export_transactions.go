package ledger

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (c *Client) ExportTransactions(
	ctx context.Context,
	input ExportTransactionsInput,
) (string, error) {
	response, err := c.service.ExportTransactions(
		ctx,
		&ledgerv1.ExportTransactionsRequest{
			UserId:   input.UserID,
			Category: input.Category,
			From:     timestamppb.New(input.From),
			To:       timestamppb.New(input.To),
		},
	)
	if err != nil {
		return "", mapError(err)
	}
	if response == nil {
		return "", ErrInvalidResponse
	}

	return response.GetCsvData(), nil
}
