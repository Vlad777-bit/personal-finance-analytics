package domain_test

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func TestNewBudget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		params  domain.NewBudgetParams
		wantErr error
	}{
		{
			name: "valid budget",
			params: domain.NewBudgetParams{
				UserID:   "user-1",
				Category: "food",
				Limit:    50000,
			},
		},
		{
			name: "empty user id",
			params: domain.NewBudgetParams{
				Category: "food",
				Limit:    50000,
			},
			wantErr: domain.ErrUserIDRequired,
		},
		{
			name: "empty category",
			params: domain.NewBudgetParams{
				UserID: "user-1",
				Limit:  50000,
			},
			wantErr: domain.ErrCategoryRequired,
		},
		{
			name: "invalid limit",
			params: domain.NewBudgetParams{
				UserID:   "user-1",
				Category: "food",
				Limit:    0,
			},
			wantErr: domain.ErrInvalidBudget,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewBudget(tt.params)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestBudget_ValidateSpending(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		spent   int64
		amount  int64
		wantErr error
	}{
		{name: "below limit", spent: 400, amount: 500},
		{name: "exact limit", spent: 400, amount: 600},
		{name: "amount exceeds limit", amount: 1001, wantErr: domain.ErrBudgetExceeded},
		{name: "total exceeds limit", spent: 400, amount: 601, wantErr: domain.ErrBudgetExceeded},
		{
			name:    "addition overflow cannot bypass limit",
			spent:   900,
			amount:  math.MaxInt64,
			wantErr: domain.ErrBudgetExceeded,
		},
	}

	budget := domain.Budget{Category: "food", Limit: 1000}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := budget.ValidateSpending(tt.spent, tt.amount)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
