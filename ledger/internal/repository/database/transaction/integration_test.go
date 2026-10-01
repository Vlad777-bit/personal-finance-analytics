//go:build integration

package transaction_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	dbpgx "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database/pgx"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	repositorypkg "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository"
	budgetrepository "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/budget"
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

	t.Run("list transactions by period and optional category", func(t *testing.T) {
		otherCategoryDate := transactionDate.AddDate(0, 0, 2)
		_, createErr := repository.Create(
			ctx,
			domain.Transaction{
				UserID:      userID,
				Amount:      3000,
				Category:    "transport",
				Description: "taxi",
				OccurredAt:  otherCategoryDate,
			},
		)
		require.NoError(t, createErr)

		from := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, 0)

		all, listErr := repository.List(ctx, repositorypkg.TransactionFilter{
			UserID: userID,
			From:   from,
			To:     to,
		})
		require.NoError(t, listErr)
		require.Len(t, all, 3)
		require.Equal(t, "transport", all[0].Category)

		food, listErr := repository.List(ctx, repositorypkg.TransactionFilter{
			UserID:   userID,
			Category: category,
			From:     from,
			To:       to,
		})
		require.NoError(t, listErr)
		require.Len(t, food, 2)
		for _, transaction := range food {
			require.Equal(t, category, transaction.Category)
		}
	})

	t.Run("list returns empty slice when transactions not found", func(t *testing.T) {
		from := time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)

		transactions, listErr := repository.List(
			ctx,
			repositorypkg.TransactionFilter{
				UserID: userID,
				From:   from,
				To:     from.AddDate(0, 1, 0),
			},
		)
		require.NoError(t, listErr)
		require.Empty(t, transactions)
		require.NotNil(t, transactions)
	})

	t.Run("concurrent transactions cannot exceed budget", func(t *testing.T) {
		const (
			concurrentUserID   = "77777777-7777-7777-7777-777777777777"
			concurrentCategory = "concurrent-budget"
		)

		testhelper.CleanupTransactions(t, ctx, client, concurrentUserID)
		testhelper.CleanupBudget(t, ctx, client, concurrentUserID, concurrentCategory)
		t.Cleanup(func() {
			testhelper.CleanupTransactions(
				t,
				context.Background(),
				client,
				concurrentUserID,
			)
			testhelper.CleanupBudget(
				t,
				context.Background(),
				client,
				concurrentUserID,
				concurrentCategory,
			)
		})

		budgetRepository := budgetrepository.New(client)
		_, upsertErr := budgetRepository.Upsert(ctx, domain.Budget{
			UserID: concurrentUserID, Category: concurrentCategory, Limit: 1000,
		})
		require.NoError(t, upsertErr)

		start := make(chan struct{})
		results := make(chan error, 2)
		for range 2 {
			go func() {
				<-start
				_, createErr := repository.CreateWithinBudget(ctx, domain.Transaction{
					UserID:     concurrentUserID,
					Amount:     600,
					Category:   concurrentCategory,
					OccurredAt: transactionDate,
				})
				results <- createErr
			}()
		}
		close(start)

		var created, rejected int
		for range 2 {
			createErr := <-results
			switch {
			case createErr == nil:
				created++
			case errors.Is(createErr, domain.ErrBudgetExceeded):
				rejected++
			default:
				require.NoError(t, createErr)
			}
		}

		require.Equal(t, 1, created)
		require.Equal(t, 1, rejected)

		from, to := monthPeriod(transactionDate)
		total, sumErr := repository.SumByCategoryAndPeriod(
			ctx,
			concurrentUserID,
			concurrentCategory,
			from,
			to,
		)
		require.NoError(t, sumErr)
		require.Equal(t, int64(600), total)

		_, createErr := repository.CreateWithinBudget(ctx, domain.Transaction{
			UserID:     concurrentUserID,
			Amount:     400,
			Category:   concurrentCategory,
			OccurredAt: transactionDate,
		})
		require.NoError(t, createErr)

		_, createErr = repository.CreateWithinBudget(ctx, domain.Transaction{
			UserID:     concurrentUserID,
			Amount:     1,
			Category:   concurrentCategory,
			OccurredAt: transactionDate,
		})
		require.ErrorIs(t, createErr, domain.ErrBudgetExceeded)

		total, sumErr = repository.SumByCategoryAndPeriod(
			ctx,
			concurrentUserID,
			concurrentCategory,
			from,
			to,
		)
		require.NoError(t, sumErr)
		require.Equal(t, int64(1000), total)
	})
}

func monthPeriod(value time.Time) (time.Time, time.Time) {
	from := time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, value.Location())

	return from, from.AddDate(0, 1, 0)
}
