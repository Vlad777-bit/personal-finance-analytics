package service

import (
	"context"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/repository"
)

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (domain.User, error)
	Login(ctx context.Context, input LoginInput) (domain.User, error)
}

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Matches(ctx context.Context, passwordHash, password string) (bool, error)
}

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

var _ AuthService = (*service)(nil)

type service struct {
	userRepository repository.UserRepository
	passwordHasher PasswordHasher
}

func New(
	userRepository repository.UserRepository,
	passwordHasher PasswordHasher,
) AuthService {
	return &service{
		userRepository: userRepository,
		passwordHasher: passwordHasher,
	}
}
