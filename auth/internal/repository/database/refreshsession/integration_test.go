//go:build integration

package refreshsession

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	dbpgx "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database/pgx"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func TestRefreshSessionRepository(t *testing.T) {
	databaseURL := os.Getenv("AUTH_DATABASE_URL")
	require.NotEmpty(t, databaseURL)
	database, err := dbpgx.New(context.Background(), databaseURL)
	require.NoError(t, err)
	t.Cleanup(database.Close)

	userID := uuid.New()
	tokenHash := "integration-refresh-session-" + uuid.NewString()
	_, err = database.Exec(context.Background(),
		"INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)",
		userID, uuid.NewString()+"@example.com", "hash")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, cleanupErr := database.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
		require.NoError(t, cleanupErr)
	})

	repository := New(database)
	require.NoError(t, repository.Create(context.Background(), domain.RefreshSession{
		TokenHash: tokenHash,
		UserID:    userID.String(),
		ExpiresAt: time.Now().Add(time.Hour),
	}))
	require.NoError(t, repository.Consume(context.Background(), tokenHash))
	require.ErrorIs(t, repository.Consume(context.Background(), tokenHash), domain.ErrRefreshSessionNotFound)
	require.NoError(t, repository.Create(context.Background(), domain.RefreshSession{
		TokenHash: tokenHash + "-expired",
		UserID:    userID.String(),
		ExpiresAt: time.Now().Add(time.Hour),
	}))
	_, err = database.Exec(context.Background(), "UPDATE refresh_sessions SET expires_at = NOW() - INTERVAL '1 hour' WHERE token_hash = $1", tokenHash+"-expired")
	require.NoError(t, err)
	require.NoError(t, repository.DeleteExpired(context.Background()))
	require.ErrorIs(t, repository.Consume(context.Background(), tokenHash+"-expired"), domain.ErrRefreshSessionNotFound)
}
