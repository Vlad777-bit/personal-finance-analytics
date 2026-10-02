package user

import (
	"context"
	"errors"
	"fmt"

	db "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func (r *Repository) GetByEmail(
	ctx context.Context,
	email string,
) (domain.User, error) {
	const query = `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, db.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("query user by email: %w", err)
	}

	return user, nil
}
