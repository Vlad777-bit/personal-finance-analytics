package grpc

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func invalidRequestError() error {
	return status.Error(codes.InvalidArgument, "request is required")
}

func invalidTimestampError(field string, err error) error {
	return status.Error(
		codes.InvalidArgument,
		fmt.Sprintf("invalid %s: %s", field, err),
	)
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, context.Canceled.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, context.DeadlineExceeded.Error())
	case errors.Is(err, domain.ErrUserIDRequired),
		errors.Is(err, domain.ErrCategoryRequired),
		errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrInvalidBudget),
		errors.Is(err, domain.ErrDateRequired),
		errors.Is(err, domain.ErrInvalidPeriod),
		errors.Is(err, domain.ErrInvalidCSV):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrBudgetNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrBudgetExceeded):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
