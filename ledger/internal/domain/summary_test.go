package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func TestBuildSummary(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	summary := domain.BuildSummary(
		"user-1",
		from,
		to,
		[]domain.CategoryReportData{
			{Category: "food", Spent: 600, BudgetLimit: 500, BudgetConfigured: true},
			{Category: "rent", BudgetLimit: 1000, BudgetConfigured: true},
			{Category: "transport", Spent: 100},
		},
	)

	require.Equal(t, int64(700), summary.TotalSpent)
	require.Equal(t, "user-1", summary.UserID)
	require.Equal(t, from, summary.From)
	require.Equal(t, to, summary.To)
	require.Equal(t, []domain.CategorySummary{
		{
			Category: "food", Spent: 600, BudgetLimit: 500,
			BudgetConfigured: true, Remaining: -100, BudgetExceeded: true,
		},
		{
			Category: "rent", BudgetLimit: 1000,
			BudgetConfigured: true, Remaining: 1000,
		},
		{Category: "transport", Spent: 100},
	}, summary.Categories)
}
