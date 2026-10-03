package service

import (
	"context"
	"fmt"
	"strings"
)

func (s *service) LogoutAll(ctx context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("user id is required")
	}
	if err := s.refreshSessions.RevokeAll(ctx, userID); err != nil {
		return fmt.Errorf("revoke all user sessions: %w", err)
	}
	return nil
}
