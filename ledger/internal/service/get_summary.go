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

	data, err := s.reportRepository.GetSummaryData(
		ctx,
		userID,
		input.From,
		input.To,
	)
	if err != nil {
		return domain.Summary{}, fmt.Errorf("get summary data: %w", err)
	}

	return domain.BuildSummary(userID, input.From, input.To, data), nil
}
