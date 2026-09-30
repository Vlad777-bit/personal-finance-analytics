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
	testUserID   = "11111111-1111-1111-1111-111111111111"
	testCategory = "food"
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

	t.Cleanup(func() {
		testhelper.CleanupBudget(
			t,
			context.Background(),
			client,
			testUserID,
			testCategory,
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
}
