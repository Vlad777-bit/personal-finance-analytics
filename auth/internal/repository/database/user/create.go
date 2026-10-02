package user

import (
	"context"
	"errors"
	"fmt"

	db "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func (r *Repository) Create(
	ctx context.Context,
	user domain.User,
) (domain.User, error) {
	const query = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at, updated_at
	`

	var createdUser domain.User
	err := r.db.QueryRow(ctx, query, user.Email, user.PasswordHash).Scan(
		&createdUser.ID,
		&createdUser.Email,
		&createdUser.PasswordHash,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)
	if errors.Is(err, db.ErrUniqueViolation) {
		return domain.User{}, domain.ErrUserAlreadyExists
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	return createdUser, nil
}
