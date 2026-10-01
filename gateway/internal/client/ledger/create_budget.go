package ledger

import (
	"context"

	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (c *Client) CreateBudget(
	ctx context.Context,
	input CreateBudgetInput,
) (Budget, error) {
	response, err := c.service.CreateBudget(
		ctx,
		&ledgerv1.CreateBudgetRequest{
			UserId:      input.UserID,
			Category:    input.Category,
			LimitAmount: input.Limit,
		},
	)
	if err != nil {
		return Budget{}, mapError(err)
	}

	return budgetFromProto(response.GetBudget())
}
