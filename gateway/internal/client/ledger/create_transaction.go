package ledger

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (c *Client) CreateTransaction(
	ctx context.Context,
	input CreateTransactionInput,
) (Transaction, error) {
	response, err := c.service.CreateTransaction(
		ctx,
		&ledgerv1.CreateTransactionRequest{
			UserId:      input.UserID,
			Amount:      input.Amount,
			Category:    input.Category,
			Description: input.Description,
			OccurredAt:  timestamppb.New(input.OccurredAt),
		},
	)
	if err != nil {
		return Transaction{}, mapError(err)
	}

	return transactionFromProto(response.GetTransaction())
}
