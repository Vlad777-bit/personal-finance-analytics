package ledger

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (c *Client) GetTransactions(
	ctx context.Context,
	input GetTransactionsInput,
) ([]Transaction, error) {
	response, err := c.service.GetTransactions(
		ctx,
		&ledgerv1.GetTransactionsRequest{
			UserId:   input.UserID,
			Category: input.Category,
			From:     timestamppb.New(input.From),
			To:       timestamppb.New(input.To),
		},
	)
	if err != nil {
		return nil, mapError(err)
	}
	if response == nil {
		return nil, ErrInvalidResponse
	}

	transactions := make([]Transaction, 0, len(response.GetTransactions()))
	for index, protoTransaction := range response.GetTransactions() {
		transaction, conversionErr := transactionFromProto(protoTransaction)
		if conversionErr != nil {
			return nil, fmt.Errorf(
				"decode transaction at index %d: %w",
				index,
				conversionErr,
			)
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}
