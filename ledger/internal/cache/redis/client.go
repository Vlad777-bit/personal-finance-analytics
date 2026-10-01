package redis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	redisdriver "github.com/redis/go-redis/v9"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/cache"
)

var _ cache.SummaryCache = (*Client)(nil)

type Client struct {
	client *redisdriver.Client
	logger *slog.Logger
}

func New(ctx context.Context, address string, logger *slog.Logger) (*Client, error) {
	client := redisdriver.NewClient(&redisdriver.Options{Addr: address})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.Join(
			fmt.Errorf("ping Redis: %w", err),
			client.Close(),
		)
	}

	return &Client{client: client, logger: logger}, nil
}

func (c *Client) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("close Redis client: %w", err)
	}

	return nil
}
