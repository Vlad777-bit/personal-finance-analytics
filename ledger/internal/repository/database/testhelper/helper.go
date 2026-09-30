//go:build integration

package testhelper

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database"
)

func RequireEnv(
	t *testing.T,
	key string,
) string {
	t.Helper()

	value := os.Getenv(key)

	require.NotEmptyf(
		t,
		value,
		"environment variable %s must be set",
		key,
	)

	return value
}

func CleanupBudget(
	t *testing.T,
	ctx context.Context,
	db database.DB,
	userID string,
	category string,
) {
	t.Helper()

	_, err := db.Exec(
		ctx,
		`
			DELETE FROM budgets
			WHERE user_id = $1
			  AND category = $2
		`,
		userID,
		category,
	)
	require.NoError(t, err)
}

func CleanupTransactions(
	t *testing.T,
	ctx context.Context,
	db database.DB,
	userID string,
) {
	t.Helper()

	_, err := db.Exec(
		ctx,
		`
			DELETE FROM transactions
			WHERE user_id = $1
		`,
		userID,
	)
	require.NoError(t, err)
}
