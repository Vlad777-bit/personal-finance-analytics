package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func (s *service) Refresh(ctx context.Context, input RefreshInput) (AccessToken, error) {
	if input.RefreshToken == "" {
		return AccessToken{}, domain.ErrInvalidRefreshToken
	}

	pair, err := s.tokenIssuer.RefreshTokens(ctx, input.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			return AccessToken{}, domain.ErrInvalidRefreshToken
		}
		return AccessToken{}, fmt.Errorf("refresh access token: %w", err)
	}
	if err := s.refreshSessions.Consume(ctx, refreshTokenHash(input.RefreshToken)); err != nil {
		return AccessToken{}, fmt.Errorf("consume refresh session: %w", err)
	}
	if err := s.refreshSessions.Create(ctx, domain.RefreshSession{
		TokenHash: refreshTokenHash(pair.RefreshToken.Value),
		UserID:    pair.UserID,
		ExpiresAt: pair.RefreshToken.ExpiresAt,
	}); err != nil {
		return AccessToken{}, fmt.Errorf("store rotated refresh session: %w", err)
	}

	return pair.AccessToken, nil
}
