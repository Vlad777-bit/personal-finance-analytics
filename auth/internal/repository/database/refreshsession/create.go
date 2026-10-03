package refreshsession

import (
	"context"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func (r *Repository) Create(ctx context.Context, session domain.RefreshSession) error {
	const query = `
		INSERT INTO refresh_sessions (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
	`

	if _, err := r.db.Exec(ctx, query, session.TokenHash, session.UserID, session.ExpiresAt); err != nil {
		return fmt.Errorf("insert refresh session: %w", err)
	}

	return nil
}
