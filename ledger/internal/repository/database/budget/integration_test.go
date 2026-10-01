//go:build integration

package budget_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	dbpgx "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database/pgx"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	budgetrepository "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/budget"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/testhelper"
)

const (
	testUserID         = "11111111-1111-1111-1111-111111111111"
	testEmptyUserID    = "11111111-1111-1111-1111-111111111112"
	testCategory       = "food"
	testSecondCategory = "transport"
)

func TestBudgetRepository(t *testing.T) {
	dsn := testhelper.RequireEnv(t, "LEDGER_DATABASE_URL")

	ctx := context.Background()

	client, err := dbpgx.New(ctx, dsn)
	require.NoError(t, err)

	t.Cleanup(client.Close)

	repository := budgetrepository.New(client)

	testhelper.CleanupBudget(
		t,
		ctx,
		client,
		testUserID,
		testCategory,
	)
	testhelper.CleanupBudget(
		t,
		ctx,
		client,
		testUserID,
		testSecondCategory,
	)

	t.Cleanup(func() {
		testhelper.CleanupBudget(
			t,
			context.Background(),
			client,
			testUserID,
			testCategory,
		)
		testhelper.CleanupBudget(
			t,
			context.Background(),
			client,
			testUserID,
			testSecondCategory,
		)
	})

	t.Run("budget not found", func(t *testing.T) {
		_, err := repository.GetByCategory(
			ctx,
			testUserID,
			testCategory,
		)

		require.ErrorIs(
			t,
			err,
			domain.ErrBudgetNotFound,
		)
	})

	var createdBudget domain.Budget

	t.Run("create budget", func(t *testing.T) {
		createdBudget, err = repository.Upsert(
			ctx,
			domain.Budget{
				UserID:   testUserID,
				Category: testCategory,
				Limit:    50000,
			},
		)
		require.NoError(t, err)

		require.NotEmpty(t, createdBudget.ID)
		require.Equal(t, testUserID, createdBudget.UserID)
		require.Equal(t, testCategory, createdBudget.Category)
		require.Equal(t, int64(50000), createdBudget.Limit)

		savedBudget, err := repository.GetByCategory(
			ctx,
			testUserID,
			testCategory,
		)
		require.NoError(t, err)

		require.Equal(t, createdBudget, savedBudget)
	})

	t.Run("update existing budget", func(t *testing.T) {
		updatedBudget, err := repository.Upsert(
			ctx,
			domain.Budget{
				UserID:   testUserID,
				Category: testCategory,
				Limit:    75000,
			},
		)
		require.NoError(t, err)

		require.Equal(t, createdBudget.ID, updatedBudget.ID)
		require.Equal(t, testUserID, updatedBudget.UserID)
		require.Equal(t, testCategory, updatedBudget.Category)
		require.Equal(t, int64(75000), updatedBudget.Limit)

		savedBudget, err := repository.GetByCategory(
			ctx,
			testUserID,
			testCategory,
		)
		require.NoError(t, err)

		require.Equal(t, updatedBudget, savedBudget)
	})

	t.Run("list budgets by user ordered by category", func(t *testing.T) {
		_, upsertErr := repository.Upsert(
			ctx,
			domain.Budget{
				UserID:   testUserID,
				Category: testSecondCategory,
				Limit:    25000,
			},
		)
		require.NoError(t, upsertErr)

		budgets, listErr := repository.ListByUser(ctx, testUserID)
		require.NoError(t, listErr)
		require.Len(t, budgets, 2)
		require.Equal(t, testCategory, budgets[0].Category)
		require.Equal(t, testSecondCategory, budgets[1].Category)
	})

	t.Run("list returns empty slice when budgets not found", func(t *testing.T) {
		budgets, listErr := repository.ListByUser(ctx, testEmptyUserID)
		require.NoError(t, listErr)
		require.Empty(t, budgets)
		require.NotNil(t, budgets)
	})
}
