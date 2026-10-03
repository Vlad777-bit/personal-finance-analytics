package service

import (
	"context"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func (s *service) Logout(ctx context.Context, input RefreshInput) error {
	if input.RefreshToken == "" {
		return domain.ErrInvalidRefreshToken
	}
	if err := s.refreshSessions.Consume(ctx, refreshTokenHash(input.RefreshToken)); err != nil {
		return fmt.Errorf("revoke refresh session: %w", err)
	}
	return nil
}
