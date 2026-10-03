package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func (s *service) Refresh(ctx context.Context, input RefreshInput) (TokenPair, error) {
	if input.RefreshToken == "" {
		return TokenPair{}, domain.ErrInvalidRefreshToken
	}

	pair, err := s.tokenIssuer.RefreshTokens(ctx, input.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			return TokenPair{}, domain.ErrInvalidRefreshToken
		}
		return TokenPair{}, fmt.Errorf("refresh access token: %w", err)
	}
	if err := s.refreshSessions.Consume(ctx, refreshTokenHash(input.RefreshToken)); err != nil {
		return TokenPair{}, fmt.Errorf("consume refresh session: %w", err)
	}
	if err := s.refreshSessions.Create(ctx, domain.RefreshSession{
		TokenHash: refreshTokenHash(pair.RefreshToken.Value),
		UserID:    pair.UserID,
		ExpiresAt: pair.RefreshToken.ExpiresAt,
	}); err != nil {
		return TokenPair{}, fmt.Errorf("store rotated refresh session: %w", err)
	}

	return pair, nil
}
