package refreshsession

import (
	"context"
	"fmt"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func (r *Repository) Consume(ctx context.Context, tokenHash string) error {
	const query = `
		UPDATE refresh_sessions
		SET revoked_at = NOW()
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
	`

	result, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("consume refresh session: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrRefreshSessionNotFound
	}
	return nil
}
