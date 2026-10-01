package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/mocks"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
)

func TestService_CreateBudget(t *testing.T) {
	t.Parallel()

	errRepository := errors.New("repository error")

	tests := []struct {
		name    string
		input   service.CreateBudgetInput
		prepare func(
			transactionRepository *mocks.TransactionRepository,
			budgetRepository *mocks.BudgetRepository,
		)
		want    domain.Budget
		wantErr error
	}{
		{
			name: "success",
			input: service.CreateBudgetInput{
				UserID:   "user-1",
				Category: "food",
				Limit:    50000,
			},
			prepare: func(
				_ *mocks.TransactionRepository,
				budgetRepository *mocks.BudgetRepository,
			) {
				budgetRepository.
					EXPECT().
					Upsert(
						mock.Anything,
						mock.MatchedBy(func(budget domain.Budget) bool {
							return budget.UserID == "user-1" &&
								budget.Category == "food" &&
								budget.Limit == 50000
						}),
					).
					Return(
						domain.Budget{
							ID:       "budget-1",
							UserID:   "user-1",
							Category: "food",
							Limit:    50000,
						},
						nil,
					)
			},
			want: domain.Budget{
				ID:       "budget-1",
				UserID:   "user-1",
				Category: "food",
				Limit:    50000,
			},
		},
		{
			name: "invalid user id",
			input: service.CreateBudgetInput{
				UserID:   "",
				Category: "food",
				Limit:    50000,
			},
			prepare: func(
				_ *mocks.TransactionRepository,
				_ *mocks.BudgetRepository,
			) {
			},
			wantErr: domain.ErrUserIDRequired,
		},
		{
			name: "invalid category",
			input: service.CreateBudgetInput{
				UserID:   "user-1",
				Category: "",
				Limit:    50000,
			},
			prepare: func(
				_ *mocks.TransactionRepository,
				_ *mocks.BudgetRepository,
			) {
			},
			wantErr: domain.ErrCategoryRequired,
		},
		{
			name: "invalid limit",
			input: service.CreateBudgetInput{
				UserID:   "user-1",
				Category: "food",
				Limit:    0,
			},
			prepare: func(
				_ *mocks.TransactionRepository,
				_ *mocks.BudgetRepository,
			) {
			},
			wantErr: domain.ErrInvalidBudget,
		},
		{
			name: "repository error",
			input: service.CreateBudgetInput{
				UserID:   "user-1",
				Category: "food",
				Limit:    50000,
			},
			prepare: func(
				_ *mocks.TransactionRepository,
				budgetRepository *mocks.BudgetRepository,
			) {
				budgetRepository.
					EXPECT().
					Upsert(
						mock.Anything,
						mock.AnythingOfType("domain.Budget"),
					).
					Return(
						domain.Budget{},
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

			ledgerService := service.New(
				transactionRepository,
				budgetRepository,
				reportRepository,
			)

			got, err := ledgerService.CreateBudget(
				context.Background(),
				tt.input,
			)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
