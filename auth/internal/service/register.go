package service

import (
	"context"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func (s *service) Register(
	ctx context.Context,
	input RegisterInput,
) (domain.User, error) {
	email, err := domain.NormalizeEmail(input.Email)
	if err != nil {
		return domain.User{}, fmt.Errorf("validate registration email: %w", err)
	}
	if err := domain.ValidatePassword(input.Password); err != nil {
		return domain.User{}, fmt.Errorf("validate registration password: %w", err)
	}

	passwordHash, err := s.passwordHasher.Hash(ctx, input.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := domain.NewUser(email, passwordHash)
	if err != nil {
		return domain.User{}, fmt.Errorf("create user entity: %w", err)
	}

	createdUser, err := s.userRepository.Create(ctx, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return createdUser, nil
}
