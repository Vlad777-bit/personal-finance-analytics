package service

import (
	"context"
	"time"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/repository"
)

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (domain.User, error)
	Login(ctx context.Context, input LoginInput) (LoginResult, error)
	Refresh(ctx context.Context, input RefreshInput) (AccessToken, error)
}

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Matches(ctx context.Context, passwordHash, password string) (bool, error)
}

type TokenIssuer interface {
	Issue(ctx context.Context, user domain.User) (AccessToken, error)
	IssueRefresh(ctx context.Context, user domain.User) (RefreshToken, error)
	Refresh(ctx context.Context, token string) (AccessToken, error)
}

type AccessToken struct {
	Value     string
	ExpiresAt time.Time
}

type RefreshToken struct {
	Value     string
	ExpiresAt time.Time
}

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type RefreshInput struct {
	RefreshToken string
}

type LoginResult struct {
	UserID       string
	Email        string
	AccessToken  AccessToken
	RefreshToken RefreshToken
}

var _ AuthService = (*service)(nil)

type service struct {
	userRepository repository.UserRepository
	passwordHasher PasswordHasher
	tokenIssuer    TokenIssuer
}

func New(
	userRepository repository.UserRepository,
	passwordHasher PasswordHasher,
	tokenIssuer TokenIssuer,
) AuthService {
	return &service{
		userRepository: userRepository,
		passwordHasher: passwordHasher,
		tokenIssuer:    tokenIssuer,
	}
}
