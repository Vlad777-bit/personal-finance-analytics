package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	redisdriver "github.com/redis/go-redis/v9"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/cache"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

const summaryKeyPrefix = "ledger:summary"

func (c *Client) Get(
	ctx context.Context,
	userID string,
	from time.Time,
	to time.Time,
) (domain.Summary, cache.Version, bool, error) {
	version, err := c.userVersion(ctx, userID)
	if err != nil {
		return domain.Summary{}, 0, false, err
	}

	key := summaryKey(userID, version, from, to)
	value, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redisdriver.Nil) {
		c.logger.Info("summary cache miss", "user_id", userID)

		return domain.Summary{}, version, false, nil
	}
	if err != nil {
		return domain.Summary{}, 0, false, fmt.Errorf("get summary cache: %w", err)
	}

	var summary domain.Summary
	if err := json.Unmarshal(value, &summary); err != nil {
		return domain.Summary{}, 0, false, fmt.Errorf("decode summary cache: %w", err)
	}

	c.logger.Info("summary cache hit", "user_id", userID)

	return summary, version, true, nil
}

func (c *Client) Set(
	ctx context.Context,
	summary domain.Summary,
	version cache.Version,
	ttl time.Duration,
) error {
	value, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("encode summary cache: %w", err)
	}

	key := summaryKey(summary.UserID, version, summary.From, summary.To)
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("set summary cache: %w", err)
	}

	return nil
}

func (c *Client) InvalidateUser(ctx context.Context, userID string) error {
	if err := c.client.Incr(ctx, versionKey(userID)).Err(); err != nil {
		return fmt.Errorf("invalidate user summary cache: %w", err)
	}

	c.logger.Info("summary cache invalidated", "user_id", userID)

	return nil
}

func (c *Client) userVersion(ctx context.Context, userID string) (cache.Version, error) {
	version, err := c.client.Get(ctx, versionKey(userID)).Int64()
	if errors.Is(err, redisdriver.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get summary cache version: %w", err)
	}

	return cache.Version(version), nil
}

func summaryKey(userID string, version cache.Version, from, to time.Time) string {
	return summaryKeyPrefix + ":data:" + userHash(userID) + ":" +
		strconv.FormatInt(int64(version), 10) + ":" +
		strconv.FormatInt(from.UnixNano(), 10) + ":" +
		strconv.FormatInt(to.UnixNano(), 10)
}

func versionKey(userID string) string {
	return summaryKeyPrefix + ":version:" + userHash(userID)
}

func userHash(userID string) string {
	sum := sha256.Sum256([]byte(userID))

	return hex.EncodeToString(sum[:])
}
