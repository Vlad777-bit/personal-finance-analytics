package jwt

import (
	"context"
	"errors"
	"testing"
	"time"

	jwtlibrary "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func TestIssuer_Issue(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)
	issuer, err := newIssuer("test-secret", "auth", time.Hour, 24*time.Hour, func() time.Time { return now })
	require.NoError(t, err)

	accessToken, err := issuer.Issue(t.Context(), domain.User{
		ID:    "user-1",
		Email: "user@example.com",
	})
	require.NoError(t, err)
	require.Equal(t, now.Add(time.Hour), accessToken.ExpiresAt)

	parsedClaims := &claims{}
	token, err := jwtlibrary.ParseWithClaims(
		accessToken.Value,
		parsedClaims,
		func(*jwtlibrary.Token) (any, error) { return []byte("test-secret"), nil },
		jwtlibrary.WithValidMethods([]string{jwtlibrary.SigningMethodHS256.Alg()}),
		jwtlibrary.WithIssuer("auth"),
		jwtlibrary.WithTimeFunc(func() time.Time { return now }),
	)
	require.NoError(t, err)
	require.True(t, token.Valid)
	require.Equal(t, "user-1", parsedClaims.Subject)
	require.Equal(t, "user@example.com", parsedClaims.Email)
	require.True(t, now.Equal(parsedClaims.IssuedAt.Time))
	require.True(t, now.Add(time.Hour).Equal(parsedClaims.ExpiresAt.Time))
}

func TestIssuer_Issue_CanceledContext(t *testing.T) {
	t.Parallel()

	issuer, err := New("test-secret", "auth", time.Hour, 24*time.Hour)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err = issuer.Issue(ctx, domain.User{ID: "user-1", Email: "user@example.com"})
	require.ErrorIs(t, err, context.Canceled)
}

func TestIssuer_Refresh(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)
	issuer, err := newIssuer("test-secret", "auth", time.Hour, 24*time.Hour, func() time.Time { return now })
	require.NoError(t, err)
	refreshToken, err := issuer.IssueRefresh(t.Context(), domain.User{ID: "user-1", Email: "user@example.com"})
	require.NoError(t, err)
	accessToken, err := issuer.Refresh(t.Context(), refreshToken.Value)
	require.NoError(t, err)
	require.Equal(t, now.Add(time.Hour), accessToken.ExpiresAt)
	_, err = issuer.Refresh(t.Context(), accessToken.Value)
	require.ErrorIs(t, err, ErrInvalidRefreshToken)
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		secret  string
		issuer  string
		ttl     time.Duration
		wantErr error
	}{
		{name: "valid", secret: "secret", issuer: "auth", ttl: time.Minute},
		{name: "empty secret", issuer: "auth", ttl: time.Minute, wantErr: ErrSecretRequired},
		{name: "blank secret", secret: " ", issuer: "auth", ttl: time.Minute, wantErr: ErrSecretRequired},
		{name: "empty issuer", secret: "secret", ttl: time.Minute, wantErr: ErrIssuerRequired},
		{name: "blank issuer", secret: "secret", issuer: " ", ttl: time.Minute, wantErr: ErrIssuerRequired},
		{name: "zero TTL", secret: "secret", issuer: "auth", wantErr: ErrInvalidTTL},
		{name: "negative TTL", secret: "secret", issuer: "auth", ttl: -time.Minute, wantErr: ErrInvalidTTL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(tt.secret, tt.issuer, tt.ttl, time.Hour)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestIssuer_Issue_RejectsExpiredToken(t *testing.T) {
	t.Parallel()

	issuedAt := time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)
	issuer, err := newIssuer("test-secret", "auth", time.Minute, time.Hour, func() time.Time { return issuedAt })
	require.NoError(t, err)
	accessToken, err := issuer.Issue(t.Context(), domain.User{ID: "user-1", Email: "user@example.com"})
	require.NoError(t, err)

	_, err = jwtlibrary.Parse(
		accessToken.Value,
		func(*jwtlibrary.Token) (any, error) { return []byte("test-secret"), nil },
		jwtlibrary.WithValidMethods([]string{jwtlibrary.SigningMethodHS256.Alg()}),
		jwtlibrary.WithTimeFunc(func() time.Time { return issuedAt.Add(2 * time.Minute) }),
	)
	require.ErrorIs(t, err, jwtlibrary.ErrTokenExpired)
	require.False(t, errors.Is(err, jwtlibrary.ErrTokenSignatureInvalid))
}
