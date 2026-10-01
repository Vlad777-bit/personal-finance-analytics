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

	tests := []struct {
		name    string
		input   service.CreateTransactionInput
		prepare func(
			transactionRepository *mocks.TransactionRepository,
			budgetRepository *mocks.BudgetRepository,
		)
		wantErr error
	}{
		{
			name: "success without budget",
			input: service.CreateTransactionInput{
				UserID:      "user-1",
				Amount:      1500,
				Category:    "food",
				Description: "lunch",
				OccurredAt:  transactionDate,
			},
			prepare: func(
				transactionRepository *mocks.TransactionRepository,
				budgetRepository *mocks.BudgetRepository,
			) {
				budgetRepository.
					EXPECT().
					GetByCategory(
						mock.Anything,
						"user-1",
						"food",
					).
					Return(
						domain.Budget{},
						domain.ErrBudgetNotFound,
					)

				transactionRepository.
					EXPECT().
					Create(
						mock.Anything,
						mock.MatchedBy(func(transaction domain.Transaction) bool {
							return transaction.UserID == "user-1" &&
								transaction.Amount == 1500 &&
								transaction.Category == "food" &&
								transaction.Description == "lunch" &&
								transaction.OccurredAt.Equal(transactionDate)
						}),
					).
					Return(
						domain.Transaction{
							ID:          "transaction-1",
							UserID:      "user-1",
							Amount:      1500,
							Category:    "food",
							Description: "lunch",
							OccurredAt:  transactionDate,
						},
						nil,
					)
			},
		},
		{
			name: "success with budget",
			input: service.CreateTransactionInput{
				UserID:     "user-1",
				Amount:     1500,
				Category:   "food",
				OccurredAt: transactionDate,
			},
			prepare: func(
				transactionRepository *mocks.TransactionRepository,
				budgetRepository *mocks.BudgetRepository,
			) {
				budgetRepository.
					EXPECT().
					GetByCategory(
						mock.Anything,
						"user-1",
						"food",
					).
					Return(
						domain.Budget{
							UserID:   "user-1",
							Category: "food",
							Limit:    10000,
						},
						nil,
					)

				transactionRepository.
					EXPECT().
					SumByCategoryAndPeriod(
						mock.Anything,
						"user-1",
						"food",
						mock.Anything,
						mock.Anything,
					).
					Return(int64(5000), nil)

				transactionRepository.
					EXPECT().
					Create(
						mock.Anything,
						mock.AnythingOfType("domain.Transaction"),
					).
					Return(
						domain.Transaction{
							ID:         "transaction-1",
							UserID:     "user-1",
							Amount:     1500,
							Category:   "food",
							OccurredAt: transactionDate,
						},
						nil,
					)
			},
		},
		{
			name: "success when budget exactly reached",
			input: service.CreateTransactionInput{
				UserID:     "user-1",
				Amount:     5000,
				Category:   "food",
				OccurredAt: transactionDate,
			},
			prepare: func(
				transactionRepository *mocks.TransactionRepository,
				budgetRepository *mocks.BudgetRepository,
			) {
				budgetRepository.
					EXPECT().
					GetByCategory(
						mock.Anything,
						"user-1",
						"food",
					).
					Return(
						domain.Budget{
							UserID:   "user-1",
							Category: "food",
							Limit:    10000,
						},
						nil,
					)

				transactionRepository.
					EXPECT().
					SumByCategoryAndPeriod(
						mock.Anything,
						"user-1",
						"food",
						mock.Anything,
						mock.Anything,
					).
					Return(int64(5000), nil)

				transactionRepository.
					EXPECT().
					Create(
						mock.Anything,
						mock.AnythingOfType("domain.Transaction"),
					).
					Return(
						domain.Transaction{
							ID:         "transaction-1",
							UserID:     "user-1",
							Amount:     5000,
							Category:   "food",
							OccurredAt: transactionDate,
						},
						nil,
					)
			},
		},
		{
			name: "budget exceeded",
			input: service.CreateTransactionInput{
				UserID:     "user-1",
				Amount:     5001,
				Category:   "food",
				OccurredAt: transactionDate,
			},
			prepare: func(
				transactionRepository *mocks.TransactionRepository,
				budgetRepository *mocks.BudgetRepository,
			) {
				budgetRepository.
					EXPECT().
					GetByCategory(
						mock.Anything,
						"user-1",
						"food",
					).
					Return(
						domain.Budget{
							UserID:   "user-1",
							Category: "food",
							Limit:    10000,
						},
						nil,
					)

				transactionRepository.
					EXPECT().
					SumByCategoryAndPeriod(
						mock.Anything,
						"user-1",
						"food",
						mock.Anything,
						mock.Anything,
					).
					Return(int64(5000), nil)
			},
			wantErr: domain.ErrBudgetExceeded,
		},
		{
			name: "invalid transaction",
			input: service.CreateTransactionInput{
				UserID:     "user-1",
				Amount:     0,
				Category:   "food",
				OccurredAt: transactionDate,
			},
			prepare: func(
				_ *mocks.TransactionRepository,
				_ *mocks.BudgetRepository,
			) {
			},
			wantErr: domain.ErrInvalidAmount,
		},
		{
			name: "budget repository error",
			input: service.CreateTransactionInput{
				UserID:     "user-1",
				Amount:     1500,
				Category:   "food",
				OccurredAt: transactionDate,
			},
			prepare: func(
				_ *mocks.TransactionRepository,
				budgetRepository *mocks.BudgetRepository,
			) {
				budgetRepository.
					EXPECT().
					GetByCategory(
						mock.Anything,
						"user-1",
						"food",
					).
					Return(
						domain.Budget{},
						errRepository,
					)
			},
			wantErr: errRepository,
		},
		{
			name: "sum repository error",
			input: service.CreateTransactionInput{
				UserID:     "user-1",
				Amount:     1500,
				Category:   "food",
				OccurredAt: transactionDate,
			},
			prepare: func(
				transactionRepository *mocks.TransactionRepository,
				budgetRepository *mocks.BudgetRepository,
			) {
				budgetRepository.
					EXPECT().
					GetByCategory(
						mock.Anything,
						"user-1",
						"food",
					).
					Return(
						domain.Budget{
							UserID:   "user-1",
							Category: "food",
							Limit:    10000,
						},
						nil,
					)

				transactionRepository.
					EXPECT().
					SumByCategoryAndPeriod(
						mock.Anything,
						"user-1",
						"food",
						mock.Anything,
						mock.Anything,
					).
					Return(int64(0), errRepository)
			},
			wantErr: errRepository,
		},
		{
			name: "create repository error",
			input: service.CreateTransactionInput{
				UserID:     "user-1",
				Amount:     1500,
				Category:   "food",
				OccurredAt: transactionDate,
			},
			prepare: func(
				transactionRepository *mocks.TransactionRepository,
				budgetRepository *mocks.BudgetRepository,
			) {
				budgetRepository.
					EXPECT().
					GetByCategory(
						mock.Anything,
						"user-1",
						"food",
					).
					Return(
						domain.Budget{},
						domain.ErrBudgetNotFound,
					)

				transactionRepository.
					EXPECT().
					Create(
						mock.Anything,
						mock.AnythingOfType("domain.Transaction"),
					).
					Return(
						domain.Transaction{},
						errRepository,
					)
			},
			wantErr: errRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			transactionRepository := mocks.NewTransactionRepository(t)
			budgetRepository := mocks.NewBudgetRepository(t)
			reportRepository := mocks.NewReportRepository(t)

			tt.prepare(
				transactionRepository,
				budgetRepository,
			)

			ledgerService := newTestLedgerService(
				transactionRepository,
				budgetRepository,
				reportRepository,
			)

			_, err := ledgerService.CreateTransaction(
				context.Background(),
				tt.input,
			)

			if tt.wantErr == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
