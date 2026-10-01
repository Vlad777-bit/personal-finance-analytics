package ledger

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func (c *Client) GetSummary(
	ctx context.Context,
	input GetSummaryInput,
) (Summary, error) {
	response, err := c.service.GetSummary(
		ctx,
		&ledgerv1.GetSummaryRequest{
			UserId: input.UserID,
			From:   timestamppb.New(input.From),
			To:     timestamppb.New(input.To),
		},
	)
	if err != nil {
		return Summary{}, mapError(err)
	}
	if response == nil || response.GetSummary() == nil {
		return Summary{}, ErrInvalidResponse
	}

	protoSummary := response.GetSummary()
	if err := validateSummaryTimestamps(protoSummary); err != nil {
		return Summary{}, err
	}

	categories := make([]CategorySummary, 0, len(protoSummary.GetCategories()))
	for index, category := range protoSummary.GetCategories() {
		if category == nil {
			return Summary{}, fmt.Errorf(
				"decode summary category at index %d: %w",
				index,
				ErrInvalidResponse,
			)
		}

		categories = append(categories, CategorySummary{
			Category:         category.GetCategory(),
			Spent:            category.GetSpent(),
			BudgetLimit:      category.GetBudgetLimit(),
			BudgetConfigured: category.GetBudgetConfigured(),
			Remaining:        category.GetRemaining(),
			BudgetExceeded:   category.GetBudgetExceeded(),
		})
	}

	return Summary{
		UserID:     protoSummary.GetUserId(),
		From:       protoSummary.GetFrom().AsTime(),
		To:         protoSummary.GetTo().AsTime(),
		TotalSpent: protoSummary.GetTotalSpent(),
		Categories: categories,
	}, nil
}

func validateSummaryTimestamps(summary *ledgerv1.Summary) error {
	if summary.GetFrom() == nil || summary.GetTo() == nil {
		return errors.Join(
			ErrInvalidResponse,
			errors.New("summary timestamps are required"),
		)
	}
	if err := summary.GetFrom().CheckValid(); err != nil {
		return fmt.Errorf("validate summary from timestamp: %w: %w", ErrInvalidResponse, err)
	}
	if err := summary.GetTo().CheckValid(); err != nil {
		return fmt.Errorf("validate summary to timestamp: %w: %w", ErrInvalidResponse, err)
	}

	return nil
}
