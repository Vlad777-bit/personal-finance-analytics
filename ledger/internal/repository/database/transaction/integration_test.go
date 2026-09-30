//go:build integration

package transaction_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	dbpgx "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database/pgx"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/testhelper"
	transactionrepository "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/transaction"
)

func TestTransactionRepository(t *testing.T) {
	dsn := testhelper.RequireEnv(t, "LEDGER_DATABASE_URL")

	ctx := context.Background()

	client, err := dbpgx.New(ctx, dsn)
	require.NoError(t, err)

	t.Cleanup(client.Close)

	repository := transactionrepository.New(client)

	userID := "22222222-2222-2222-2222-222222222222"
	category := "food"

	testhelper.CleanupTransactions(
		t,
		ctx,
		client,
		userID,
	)

	t.Cleanup(func() {
		testhelper.CleanupTransactions(
			t,
			context.Background(),
			client,
			userID,
		)
	})

	transactionDate := time.Date(
		2026,
		time.September,
		15,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	t.Run("create transaction", func(t *testing.T) {
		transaction, createErr := repository.Create(
			ctx,
			domain.Transaction{
				UserID:      userID,
				Amount:      1500,
				Category:    category,
				Description: "lunch",
				OccurredAt:  transactionDate,
			},
		)
		require.NoError(t, createErr)

		require.NotEmpty(t, transaction.ID)
		require.Equal(t, userID, transaction.UserID)
		require.Equal(t, int64(1500), transaction.Amount)
		require.Equal(t, category, transaction.Category)
		require.Equal(t, "lunch", transaction.Description)
		require.True(t, transaction.OccurredAt.Equal(transactionDate))
		require.False(t, transaction.CreatedAt.IsZero())
	})

	t.Run("sum by category and period", func(t *testing.T) {
		secondTransactionDate := transactionDate.AddDate(0, 0, 1)

		_, createErr := repository.Create(
			ctx,
			domain.Transaction{
				UserID:      userID,
				Amount:      2500,
				Category:    category,
				Description: "dinner",
				OccurredAt:  secondTransactionDate,
			},
		)
		require.NoError(t, createErr)

		from := time.Date(
			2026,
			time.September,
			1,
			0,
			0,
			0,
			0,
			time.UTC,
		)

		to := from.AddDate(0, 1, 0)

		total, sumErr := repository.SumByCategoryAndPeriod(
			ctx,
			userID,
			category,
			from,
			to,
		)
		require.NoError(t, sumErr)

		require.Equal(t, int64(4000), total)
	})

	t.Run("sum returns zero when transactions not found", func(t *testing.T) {
		from := time.Date(
			2027,
			time.January,
			1,
			0,
			0,
			0,
			0,
			time.UTC,
		)

		to := from.AddDate(0, 1, 0)

		total, sumErr := repository.SumByCategoryAndPeriod(
			ctx,
			userID,
			category,
			from,
			to,
		)
		require.NoError(t, sumErr)

		require.Zero(t, total)
	})
}
