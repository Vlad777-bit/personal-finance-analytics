//go:build integration

package report_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	dbpgx "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database/pgx"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	budgetrepository "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/budget"
	reportrepository "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/report"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/testhelper"
	transactionrepository "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/transaction"
)

func TestReportRepository_GetSummaryData(t *testing.T) {
	dsn := testhelper.RequireEnv(t, "LEDGER_DATABASE_URL")
	ctx := context.Background()
	client, err := dbpgx.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(client.Close)

	const userID = "44444444-4444-4444-4444-444444444444"
	categories := []string{"food", "rent", "transport"}
	testhelper.CleanupTransactions(t, ctx, client, userID)
	for _, category := range categories {
		testhelper.CleanupBudget(t, ctx, client, userID, category)
	}
	t.Cleanup(func() {
		cleanupContext := context.Background()
		testhelper.CleanupTransactions(t, cleanupContext, client, userID)
		for _, category := range categories {
			testhelper.CleanupBudget(t, cleanupContext, client, userID, category)
		}
	})

	budgetRepository := budgetrepository.New(client)
	transactionRepository := transactionrepository.New(client)
	reportRepository := reportrepository.New(client)
	from := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	_, err = budgetRepository.Upsert(ctx, domain.Budget{
		UserID: userID, Category: "food", Limit: 500,
	})
	require.NoError(t, err)
	_, err = budgetRepository.Upsert(ctx, domain.Budget{
		UserID: userID, Category: "rent", Limit: 1000,
	})
	require.NoError(t, err)

	transactions := []domain.Transaction{
		{UserID: userID, Category: "food", Amount: 600, OccurredAt: from.Add(time.Hour)},
		{UserID: userID, Category: "transport", Amount: 100, OccurredAt: from.Add(2 * time.Hour)},
		{UserID: userID, Category: "food", Amount: 900, OccurredAt: to},
	}
	for _, transaction := range transactions {
		_, err = transactionRepository.Create(ctx, transaction)
		require.NoError(t, err)
	}

	data, err := reportRepository.GetSummaryData(ctx, userID, from, to)
	require.NoError(t, err)
	require.Equal(t, []domain.CategoryReportData{
		{Category: "food", Spent: 600, BudgetLimit: 500, BudgetConfigured: true},
		{Category: "rent", BudgetLimit: 1000, BudgetConfigured: true},
		{Category: "transport", Spent: 100},
	}, data)
}
