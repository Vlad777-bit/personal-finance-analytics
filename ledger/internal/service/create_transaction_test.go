package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/mocks"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
)

func TestService_CreateTransaction(t *testing.T) {
	t.Parallel()

	transactionDate := time.Date(
		2026,
		time.September,
		29,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	errRepository := errors.New("repository error")
	validInput := service.CreateTransactionInput{
		UserID:      "user-1",
		Amount:      1500,
		Category:    "food",
		Description: "lunch",
		OccurredAt:  transactionDate,
	}

	tests := []struct {
		name    string
		input   service.CreateTransactionInput
		prepare func(*mocks.TransactionRepository)
		wantErr error
	}{
		{
			name:  "success",
			input: validInput,
			prepare: func(transactionRepository *mocks.TransactionRepository) {
				transactionRepository.EXPECT().
					CreateWithinBudget(
						mock.Anything,
						mock.MatchedBy(func(transaction domain.Transaction) bool {
							return transaction.UserID == "user-1" &&
								transaction.Amount == 1500 &&
								transaction.Category == "food" &&
								transaction.Description == "lunch" &&
								transaction.OccurredAt.Equal(transactionDate)
						}),
					).
					Return(domain.Transaction{
						ID:          "transaction-1",
						UserID:      "user-1",
						Amount:      1500,
						Category:    "food",
						Description: "lunch",
						OccurredAt:  transactionDate,
					}, nil)
			},
		},
		{
			name: "invalid transaction",
			input: service.CreateTransactionInput{
				UserID: "user-1", Amount: 0, Category: "food", OccurredAt: transactionDate,
			},
			prepare: func(*mocks.TransactionRepository) {},
			wantErr: domain.ErrInvalidAmount,
		},
		{
			name:  "budget exceeded",
			input: validInput,
			prepare: func(transactionRepository *mocks.TransactionRepository) {
				transactionRepository.EXPECT().
					CreateWithinBudget(mock.Anything, mock.AnythingOfType("domain.Transaction")).
					Return(domain.Transaction{}, domain.ErrBudgetExceeded)
			},
			wantErr: domain.ErrBudgetExceeded,
		},
		{
			name:  "repository error",
			input: validInput,
			prepare: func(transactionRepository *mocks.TransactionRepository) {
				transactionRepository.EXPECT().
					CreateWithinBudget(mock.Anything, mock.AnythingOfType("domain.Transaction")).
					Return(domain.Transaction{}, errRepository)
			},
			wantErr: errRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			transactionRepository := mocks.NewTransactionRepository(t)
			tt.prepare(transactionRepository)
			ledgerService := newTestLedgerService(
				transactionRepository,
				mocks.NewBudgetRepository(t),
				mocks.NewReportRepository(t),
			)

			_, err := ledgerService.CreateTransaction(context.Background(), tt.input)
			if tt.wantErr == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
