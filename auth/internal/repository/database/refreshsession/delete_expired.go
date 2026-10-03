package refreshsession

import (
	"context"
	"fmt"
)

func (r *Repository) DeleteExpired(ctx context.Context) error {
	const query = `
		DELETE FROM refresh_sessions
		WHERE expires_at <= NOW()
	`

	if _, err := r.db.Exec(ctx, query); err != nil {
		return fmt.Errorf("delete expired refresh sessions: %w", err)
	}

	return nil
}
