package grpc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func TestMapServiceError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "canceled", err: context.Canceled, code: codes.Canceled},
		{name: "deadline", err: context.DeadlineExceeded, code: codes.DeadlineExceeded},
		{name: "invalid user", err: domain.ErrUserIDRequired, code: codes.InvalidArgument},
		{name: "invalid category", err: domain.ErrCategoryRequired, code: codes.InvalidArgument},
		{name: "invalid amount", err: domain.ErrInvalidAmount, code: codes.InvalidArgument},
		{name: "invalid budget", err: domain.ErrInvalidBudget, code: codes.InvalidArgument},
		{name: "invalid date", err: domain.ErrDateRequired, code: codes.InvalidArgument},
		{name: "invalid period", err: domain.ErrInvalidPeriod, code: codes.InvalidArgument},
		{name: "budget not found", err: domain.ErrBudgetNotFound, code: codes.NotFound},
		{name: "budget exceeded", err: domain.ErrBudgetExceeded, code: codes.FailedPrecondition},
		{name: "internal", err: errors.New("database error"), code: codes.Internal},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := mapServiceError(fmt.Errorf("service operation: %w", test.err))
			require.Equal(t, test.code, status.Code(err))
		})
	}
}
