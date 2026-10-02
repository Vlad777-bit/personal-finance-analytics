package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func (s *service) Login(
	ctx context.Context,
	input LoginInput,
) (domain.User, error) {
	email, err := domain.NormalizeEmail(input.Email)
	if err != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	if err := domain.ValidatePassword(input.Password); err != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}

	user, err := s.userRepository.GetByEmail(ctx, email)
	if errors.Is(err, domain.ErrUserNotFound) {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}

	matches, err := s.passwordHasher.Matches(ctx, user.PasswordHash, input.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("compare password: %w", err)
	}
	if !matches {
		return domain.User{}, domain.ErrInvalidCredentials
	}

	return user, nil
}
