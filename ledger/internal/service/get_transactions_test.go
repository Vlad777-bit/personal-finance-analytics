package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/mocks"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
)

func TestService_GetTransactions(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	errRepository := errors.New("repository error")

	tests := []struct {
		name    string
		input   service.GetTransactionsInput
		prepare func(*mocks.TransactionRepository)
		want    []domain.Transaction
		wantErr error
	}{
		{
			name: "success with normalized filter",
			input: service.GetTransactionsInput{
				UserID:   " user-1 ",
				Category: " food ",
				From:     from,
				To:       to,
			},
			prepare: func(transactionRepository *mocks.TransactionRepository) {
				transactionRepository.EXPECT().
					List(mock.Anything, repository.TransactionFilter{
						UserID:   "user-1",
						Category: "food",
						From:     from,
						To:       to,
					}).
					Return([]domain.Transaction{{ID: "transaction-1"}}, nil)
			},
			want: []domain.Transaction{{ID: "transaction-1"}},
		},
		{
			name:    "user id required",
			input:   service.GetTransactionsInput{From: from, To: to},
			prepare: func(_ *mocks.TransactionRepository) {},
			wantErr: domain.ErrUserIDRequired,
		},
		{
			name:    "period start required",
			input:   service.GetTransactionsInput{UserID: "user-1", To: to},
			prepare: func(_ *mocks.TransactionRepository) {},
			wantErr: domain.ErrInvalidPeriod,
		},
		{
			name:    "period end required",
			input:   service.GetTransactionsInput{UserID: "user-1", From: from},
			prepare: func(_ *mocks.TransactionRepository) {},
			wantErr: domain.ErrInvalidPeriod,
		},
		{
			name:    "period must be increasing",
			input:   service.GetTransactionsInput{UserID: "user-1", From: to, To: from},
			prepare: func(_ *mocks.TransactionRepository) {},
			wantErr: domain.ErrInvalidPeriod,
		},
		{
			name:  "repository error",
			input: service.GetTransactionsInput{UserID: "user-1", From: from, To: to},
			prepare: func(transactionRepository *mocks.TransactionRepository) {
				transactionRepository.EXPECT().
					List(mock.Anything, repository.TransactionFilter{
						UserID: "user-1",
						From:   from,
						To:     to,
					}).
					Return(nil, errRepository)
			},
			wantErr: errRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			transactionRepository := mocks.NewTransactionRepository(t)
			budgetRepository := mocks.NewBudgetRepository(t)
			tt.prepare(transactionRepository)

			ledgerService := service.New(transactionRepository, budgetRepository)
			got, err := ledgerService.GetTransactions(t.Context(), tt.input)

			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}
