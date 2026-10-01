package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

func (s *service) GetSummary(
	ctx context.Context,
	input GetSummaryInput,
) (domain.Summary, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return domain.Summary{}, domain.ErrUserIDRequired
	}
	if input.From.IsZero() || input.To.IsZero() || !input.From.Before(input.To) {
		return domain.Summary{}, domain.ErrInvalidPeriod
	}

	summary, version, hit, err := s.summaryCache.Get(ctx, userID, input.From, input.To)
	cacheAvailable := err == nil
	if err != nil {
		s.logger.Warn("get summary from cache", "error", err)
	} else if hit {
		return summary, nil
	}

	data, err := s.reportRepository.GetSummaryData(
		ctx,
		userID,
		input.From,
		input.To,
	)
	if err != nil {
		return domain.Summary{}, fmt.Errorf("get summary data: %w", err)
	}

	summary = domain.BuildSummary(userID, input.From, input.To, data)
	if cacheAvailable {
		if err := s.summaryCache.Set(ctx, summary, version, s.summaryCacheTTL); err != nil {
			s.logger.Warn("store summary in cache", "error", err)
		}
	}

	return summary, nil
}
