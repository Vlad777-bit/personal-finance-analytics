package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func TestNewTransaction(t *testing.T) {
	t.Parallel()

	transactionDate := time.Date(2026, time.September, 29, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		params  domain.NewTransactionParams
		wantErr error
	}{
		{
			name: "valid transaction",
			params: domain.NewTransactionParams{
				UserID:      "user-1",
				Amount:      1500,
				Category:    "food",
				Description: "lunch",
				OccurredAt:  transactionDate,
			},
		},
		{
			name: "empty user id",
			params: domain.NewTransactionParams{
				Amount:     1500,
				Category:   "food",
				OccurredAt: transactionDate,
			},
			wantErr: domain.ErrUserIDRequired,
		},
		{
			name: "zero amount",
			params: domain.NewTransactionParams{
				UserID:     "user-1",
				Category:   "food",
				OccurredAt: transactionDate,
			},
			wantErr: domain.ErrInvalidAmount,
		},
		{
			name: "empty category",
			params: domain.NewTransactionParams{
				UserID:     "user-1",
				Amount:     1500,
				OccurredAt: transactionDate,
			},
			wantErr: domain.ErrCategoryRequired,
		},
		{
			name: "empty date",
			params: domain.NewTransactionParams{
				UserID:   "user-1",
				Amount:   1500,
				Category: "food",
			},
			wantErr: domain.ErrDateRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewTransaction(tt.params)

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
