package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func (s *service) Login(
	ctx context.Context,
	input LoginInput,
) (LoginResult, error) {
	email, err := domain.NormalizeEmail(input.Email)
	if err != nil {
		return LoginResult{}, domain.ErrInvalidCredentials
	}
	if err := domain.ValidatePassword(input.Password); err != nil {
		return LoginResult{}, domain.ErrInvalidCredentials
	}

	user, err := s.userRepository.GetByEmail(ctx, email)
	if errors.Is(err, domain.ErrUserNotFound) {
		return LoginResult{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, fmt.Errorf("get user by email: %w", err)
	}

	matches, err := s.passwordHasher.Matches(ctx, user.PasswordHash, input.Password)
	if err != nil {
		return LoginResult{}, fmt.Errorf("compare password: %w", err)
	}
	if !matches {
		return LoginResult{}, domain.ErrInvalidCredentials
	}

	accessToken, err := s.tokenIssuer.Issue(ctx, user)
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue access token: %w", err)
	}
	refreshToken, err := s.tokenIssuer.IssueRefresh(ctx, user)
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue refresh token: %w", err)
	}
	if err := s.refreshSessions.Create(ctx, domain.RefreshSession{
		TokenHash: refreshTokenHash(refreshToken.Value),
		UserID:    user.ID,
		ExpiresAt: refreshToken.ExpiresAt,
	}); err != nil {
		return LoginResult{}, fmt.Errorf("store refresh session: %w", err)
	}

	return LoginResult{
		UserID:       user.ID,
		Email:        user.Email,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func refreshTokenHash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", hash[:])
}
