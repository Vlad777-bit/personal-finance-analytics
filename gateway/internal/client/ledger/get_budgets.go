package ledger

import (
	"context"
	"fmt"

	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (c *Client) GetBudgets(
	ctx context.Context,
	input GetBudgetsInput,
) ([]Budget, error) {
	response, err := c.service.GetBudgets(
		ctx,
		&ledgerv1.GetBudgetsRequest{UserId: input.UserID},
	)
	if err != nil {
		return nil, mapError(err)
	}
	if response == nil {
		return nil, ErrInvalidResponse
	}

	budgets := make([]Budget, 0, len(response.GetBudgets()))
	for index, protoBudget := range response.GetBudgets() {
		budget, conversionErr := budgetFromProto(protoBudget)
		if conversionErr != nil {
			return nil, fmt.Errorf(
				"decode budget at index %d: %w",
				index,
				conversionErr,
			)
		}

		budgets = append(budgets, budget)
	}

	return budgets, nil
}
