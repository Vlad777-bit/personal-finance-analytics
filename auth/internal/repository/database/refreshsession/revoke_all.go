package refreshsession

import (
	"context"
	"fmt"
)

func (r *Repository) RevokeAll(ctx context.Context, userID string) error {
	const query = `
		UPDATE refresh_sessions
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`
	if _, err := r.db.Exec(ctx, query, userID); err != nil {
		return fmt.Errorf("revoke all refresh sessions: %w", err)
	}
	return nil
}
