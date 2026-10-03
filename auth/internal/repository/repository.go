package repository

import (
	"context"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
}

type RefreshSessionRepository interface {
	Create(ctx context.Context, session domain.RefreshSession) error
	Consume(ctx context.Context, tokenHash string) error
}
