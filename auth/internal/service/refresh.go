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

	token, err := s.tokenIssuer.Refresh(ctx, input.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			return AccessToken{}, domain.ErrInvalidRefreshToken
		}
		return AccessToken{}, fmt.Errorf("refresh access token: %w", err)
	}

	return token, nil
}
