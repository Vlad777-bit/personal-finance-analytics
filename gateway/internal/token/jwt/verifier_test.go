package jwt

import (
	"context"
	"testing"
	"time"

	jwtlibrary "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestVerifier_Verify(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		secret    string
		method    jwtlibrary.SigningMethod
		claims    claims
		wantError bool
	}{
		{
			name:   "valid",
			secret: "test-secret",
			method: jwtlibrary.SigningMethodHS256,
			claims: claims{
				Email: "user@example.com",
				RegisteredClaims: jwtlibrary.RegisteredClaims{
					Issuer: "auth", Subject: "user-1",
					ExpiresAt: jwtlibrary.NewNumericDate(now.Add(time.Minute)),
				},
			},
		},
		{
			name:      "wrong signature",
			secret:    "wrong-secret",
			method:    jwtlibrary.SigningMethodHS256,
			claims:    validClaims(now),
			wantError: true,
		},
		{
			name:      "wrong algorithm",
			secret:    "test-secret",
			method:    jwtlibrary.SigningMethodHS384,
			claims:    validClaims(now),
			wantError: true,
		},
		{
			name:   "expired",
			secret: "test-secret",
			method: jwtlibrary.SigningMethodHS256,
			claims: claims{
				Email: "user@example.com",
				RegisteredClaims: jwtlibrary.RegisteredClaims{
					Issuer: "auth", Subject: "user-1",
					ExpiresAt: jwtlibrary.NewNumericDate(now.Add(-time.Second)),
				},
			},
			wantError: true,
		},
		{name: "missing expiry", secret: "test-secret", method: jwtlibrary.SigningMethodHS256, claims: claims{Email: "user@example.com", RegisteredClaims: jwtlibrary.RegisteredClaims{Issuer: "auth", Subject: "user-1"}}, wantError: true},
		{name: "wrong issuer", secret: "test-secret", method: jwtlibrary.SigningMethodHS256, claims: claims{Email: "user@example.com", RegisteredClaims: jwtlibrary.RegisteredClaims{Issuer: "other", Subject: "user-1", ExpiresAt: jwtlibrary.NewNumericDate(now.Add(time.Minute))}}, wantError: true},
		{name: "missing subject", secret: "test-secret", method: jwtlibrary.SigningMethodHS256, claims: claims{Email: "user@example.com", RegisteredClaims: jwtlibrary.RegisteredClaims{Issuer: "auth", ExpiresAt: jwtlibrary.NewNumericDate(now.Add(time.Minute))}}, wantError: true},
		{name: "missing email", secret: "test-secret", method: jwtlibrary.SigningMethodHS256, claims: claims{RegisteredClaims: jwtlibrary.RegisteredClaims{Issuer: "auth", Subject: "user-1", ExpiresAt: jwtlibrary.NewNumericDate(now.Add(time.Minute))}}, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			signedToken, err := jwtlibrary.NewWithClaims(tt.method, tt.claims).SignedString([]byte(tt.secret))
			require.NoError(t, err)
			verifier, err := newVerifier("test-secret", "auth", func() time.Time { return now })
			require.NoError(t, err)

			identity, err := verifier.Verify(t.Context(), signedToken)
			if tt.wantError {
				require.ErrorIs(t, err, ErrInvalidToken)
				require.Empty(t, identity)

				return
			}

			require.NoError(t, err)
			require.Equal(t, "user-1", identity.UserID)
			require.Equal(t, "user@example.com", identity.Email)
		})
	}
}

func TestVerifier_VerifyCanceledContext(t *testing.T) {
	t.Parallel()

	verifier, err := New("test-secret", "auth")
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err = verifier.Verify(ctx, "token")
	require.ErrorIs(t, err, context.Canceled)
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		secret  string
		issuer  string
		wantErr error
	}{
		{name: "valid", secret: "secret", issuer: "auth"},
		{name: "missing secret", issuer: "auth", wantErr: ErrSecretRequired},
		{name: "blank secret", secret: " ", issuer: "auth", wantErr: ErrSecretRequired},
		{name: "missing issuer", secret: "secret", wantErr: ErrIssuerRequired},
		{name: "blank issuer", secret: "secret", issuer: " ", wantErr: ErrIssuerRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(tt.secret, tt.issuer)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func validClaims(now time.Time) claims {
	return claims{
		Email: "user@example.com",
		RegisteredClaims: jwtlibrary.RegisteredClaims{
			Issuer: "auth", Subject: "user-1",
			ExpiresAt: jwtlibrary.NewNumericDate(now.Add(time.Minute)),
		},
	}
}
