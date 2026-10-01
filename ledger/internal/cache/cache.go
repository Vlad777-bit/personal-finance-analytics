package cache

import (
	"context"
	"time"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

type Version int64

type SummaryCache interface {
	Get(
		ctx context.Context,
		userID string,
		from time.Time,
		to time.Time,
	) (domain.Summary, Version, bool, error)

	Set(
		ctx context.Context,
		summary domain.Summary,
		version Version,
		ttl time.Duration,
	) error

	InvalidateUser(ctx context.Context, userID string) error
}
