//go:build integration

package user_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	dbpgx "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database/pgx"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
	userrepository "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/repository/database/user"
)

func TestUserRepository(t *testing.T) {
	dsn := os.Getenv("AUTH_DATABASE_URL")
	require.NotEmpty(t, dsn, "AUTH_DATABASE_URL must be set")

	ctx := context.Background()
	client, err := dbpgx.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(client.Close)

	const email = "repository-integration@example.com"
	cleanupUser(t, ctx, client, email)
	t.Cleanup(func() { cleanupUser(t, context.Background(), client, email) })

	repository := userrepository.New(client)
	created, err := repository.Create(ctx, domain.User{
		Email: email, PasswordHash: "password-hash",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	require.Equal(t, email, created.Email)
	require.Equal(t, "password-hash", created.PasswordHash)
	require.False(t, created.CreatedAt.IsZero())
	require.False(t, created.UpdatedAt.IsZero())

	found, err := repository.GetByEmail(ctx, email)
	require.NoError(t, err)
	require.Equal(t, created, found)

	_, err = repository.Create(ctx, domain.User{
		Email: email, PasswordHash: "another-hash",
	})
	require.ErrorIs(t, err, domain.ErrUserAlreadyExists)

	_, err = repository.GetByEmail(ctx, "missing@example.com")
	require.ErrorIs(t, err, domain.ErrUserNotFound)
}

func cleanupUser(t *testing.T, ctx context.Context, client *dbpgx.Client, email string) {
	t.Helper()

	_, err := client.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
	require.NoError(t, err)
}
