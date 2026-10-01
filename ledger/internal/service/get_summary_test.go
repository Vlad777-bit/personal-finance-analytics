package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/mocks"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
)

func TestService_GetSummary(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	errRepository := errors.New("repository error")

	tests := []struct {
		name    string
		input   service.GetSummaryInput
		prepare func(*mocks.ReportRepository)
		want    domain.Summary
		wantErr error
	}{
		{
			name: "success",
			input: service.GetSummaryInput{
				UserID: " user-1 ", From: from, To: to,
			},
			prepare: func(reportRepository *mocks.ReportRepository) {
				reportRepository.EXPECT().
					GetSummaryData(mock.Anything, "user-1", from, to).
					Return([]domain.CategoryReportData{
						{
							Category: "food", Spent: 600, BudgetLimit: 500,
							BudgetConfigured: true,
						},
					}, nil)
			},
			want: domain.Summary{
				UserID: "user-1", From: from, To: to, TotalSpent: 600,
				Categories: []domain.CategorySummary{
					{
						Category: "food", Spent: 600, BudgetLimit: 500,
						BudgetConfigured: true, Remaining: -100, BudgetExceeded: true,
					},
				},
			},
		},
		{
			name:  "empty summary",
			input: service.GetSummaryInput{UserID: "user-1", From: from, To: to},
			prepare: func(reportRepository *mocks.ReportRepository) {
				reportRepository.EXPECT().
					GetSummaryData(mock.Anything, "user-1", from, to).
					Return([]domain.CategoryReportData{}, nil)
			},
			want: domain.Summary{
				UserID: "user-1", From: from, To: to,
				Categories: []domain.CategorySummary{},
			},
		},
		{
			name:    "user id required",
			input:   service.GetSummaryInput{From: from, To: to},
			prepare: func(_ *mocks.ReportRepository) {},
			wantErr: domain.ErrUserIDRequired,
		},
		{
			name:    "period start required",
			input:   service.GetSummaryInput{UserID: "user-1", To: to},
			prepare: func(_ *mocks.ReportRepository) {},
			wantErr: domain.ErrInvalidPeriod,
		},
		{
			name:    "period end required",
			input:   service.GetSummaryInput{UserID: "user-1", From: from},
			prepare: func(_ *mocks.ReportRepository) {},
			wantErr: domain.ErrInvalidPeriod,
		},
		{
			name:    "period must be increasing",
			input:   service.GetSummaryInput{UserID: "user-1", From: to, To: from},
			prepare: func(_ *mocks.ReportRepository) {},
			wantErr: domain.ErrInvalidPeriod,
		},
		{
			name:  "repository error",
			input: service.GetSummaryInput{UserID: "user-1", From: from, To: to},
			prepare: func(reportRepository *mocks.ReportRepository) {
				reportRepository.EXPECT().
					GetSummaryData(mock.Anything, "user-1", from, to).
					Return(nil, errRepository)
			},
			wantErr: errRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reportRepository := mocks.NewReportRepository(t)
			tt.prepare(reportRepository)
			ledgerService := service.New(
				mocks.NewTransactionRepository(t),
				mocks.NewBudgetRepository(t),
				reportRepository,
			)

			got, err := ledgerService.GetSummary(t.Context(), tt.input)

			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}
