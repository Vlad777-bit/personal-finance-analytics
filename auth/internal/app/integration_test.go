//go:build integration

package app

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/service"
)

func TestAuthApplication(t *testing.T) {
	const email = "auth-app-integration@example.com"

	databaseURL := os.Getenv("AUTH_DATABASE_URL")
	require.NotEmpty(t, databaseURL, "AUTH_DATABASE_URL must be set")

	application, err := New(
		t.Context(),
		databaseURL,
		4,
		"0123456789abcdef0123456789abcdef",
		"auth-integration-test",
		15*time.Minute,
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		require.NoError(t, application.Shutdown(shutdownContext))
	})

	cleanupUser(t, application.database, email)
	t.Cleanup(func() { cleanupUser(t, application.database, email) })

	createdUser, err := application.authService.Register(t.Context(), service.RegisterInput{
		Email:    email,
		Password: "secure-password",
	})
	require.NoError(t, err)
	require.NotEmpty(t, createdUser.ID)
	require.NotEqual(t, "secure-password", createdUser.PasswordHash)

	loginResult, err := application.authService.Login(t.Context(), service.LoginInput{
		Email:    email,
		Password: "secure-password",
	})
	require.NoError(t, err)
	require.Equal(t, createdUser.ID, loginResult.UserID)
	require.Equal(t, email, loginResult.Email)
	require.NotEmpty(t, loginResult.AccessToken.Value)
	require.WithinDuration(t, time.Now().Add(15*time.Minute), loginResult.AccessToken.ExpiresAt, time.Second)

	_, err = application.authService.Login(t.Context(), service.LoginInput{
		Email:    email,
		Password: "wrong-password",
	})
	require.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func cleanupUser(t *testing.T, db database.DB, email string) {
	t.Helper()

	_, err := db.Exec(
		context.Background(),
		"DELETE FROM users WHERE email = $1",
		email,
	)
	require.NoError(t, err)
}
