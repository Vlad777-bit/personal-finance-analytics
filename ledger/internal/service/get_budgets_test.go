package service_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/mocks"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
)

func TestService_GetBudgets(t *testing.T) {
	t.Parallel()

	errRepository := errors.New("repository error")
	tests := []struct {
		name    string
		input   service.GetBudgetsInput
		prepare func(*mocks.BudgetRepository)
		want    []domain.Budget
		wantErr error
	}{
		{
			name:  "success with normalized user id",
			input: service.GetBudgetsInput{UserID: " user-1 "},
			prepare: func(budgetRepository *mocks.BudgetRepository) {
				budgetRepository.EXPECT().
					ListByUser(mock.Anything, "user-1").
					Return([]domain.Budget{{ID: "budget-1"}}, nil)
			},
			want: []domain.Budget{{ID: "budget-1"}},
		},
		{
			name:    "user id required",
			input:   service.GetBudgetsInput{UserID: " "},
			prepare: func(_ *mocks.BudgetRepository) {},
			wantErr: domain.ErrUserIDRequired,
		},
		{
			name:  "repository error",
			input: service.GetBudgetsInput{UserID: "user-1"},
			prepare: func(budgetRepository *mocks.BudgetRepository) {
				budgetRepository.EXPECT().
					ListByUser(mock.Anything, "user-1").
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
			reportRepository := mocks.NewReportRepository(t)
			tt.prepare(budgetRepository)

			ledgerService := service.New(
				transactionRepository,
				budgetRepository,
				reportRepository,
			)
			got, err := ledgerService.GetBudgets(t.Context(), tt.input)

			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}
