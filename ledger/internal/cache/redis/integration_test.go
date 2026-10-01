//go:build integration

package redis_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	rediscache "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/cache/redis"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func TestSummaryCache(t *testing.T) {
	address := redisAddress()

	ctx := context.Background()
	client, err := rediscache.New(
		ctx,
		address,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	from := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	summary := domain.Summary{
		UserID: "55555555-5555-5555-5555-555555555555",
		From:   from,
		To:     to,
		Categories: []domain.CategorySummary{
			{Category: "food", Spent: 100},
		},
		TotalSpent: 100,
	}
	require.NoError(t, client.InvalidateUser(ctx, summary.UserID))

	result, version, hit, err := client.Get(ctx, summary.UserID, from, to)
	require.NoError(t, err)
	require.False(t, hit)
	require.Empty(t, result)

	require.NoError(t, client.Set(ctx, summary, version, time.Minute))
	result, version, hit, err = client.Get(ctx, summary.UserID, from, to)
	require.NoError(t, err)
	require.True(t, hit)
	require.Equal(t, summary, result)

	require.NoError(t, client.InvalidateUser(ctx, summary.UserID))
	require.NoError(t, client.Set(ctx, summary, version, time.Minute))
	_, version, hit, err = client.Get(ctx, summary.UserID, from, to)
	require.NoError(t, err)
	require.False(t, hit)

	require.NoError(t, client.Set(ctx, summary, version, 20*time.Millisecond))
	require.Eventually(t, func() bool {
		_, _, cacheHit, cacheErr := client.Get(ctx, summary.UserID, from, to)

		return cacheErr == nil && !cacheHit
	}, time.Second, 20*time.Millisecond)
}

func redisAddress() string {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	return net.JoinHostPort(host, port)
}
