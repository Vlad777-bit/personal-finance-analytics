package bcrypt_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	bcryptdriver "golang.org/x/crypto/bcrypt"

	passwordbcrypt "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/password/bcrypt"
)

func TestHasher(t *testing.T) {
	t.Parallel()

	hasher, err := passwordbcrypt.New(bcryptdriver.MinCost)
	require.NoError(t, err)

	passwordHash, err := hasher.Hash(t.Context(), "correct-password")
	require.NoError(t, err)
	require.NotEqual(t, "correct-password", passwordHash)

	matches, err := hasher.Matches(t.Context(), passwordHash, "correct-password")
	require.NoError(t, err)
	require.True(t, matches)

	matches, err = hasher.Matches(t.Context(), passwordHash, "wrong-password")
	require.NoError(t, err)
	require.False(t, matches)
}

func TestHasherErrors(t *testing.T) {
	t.Parallel()

	hasher, err := passwordbcrypt.New(bcryptdriver.MinCost)
	require.NoError(t, err)
	tooLongPassword := strings.Repeat("a", 73)

	tests := []struct {
		name    string
		call    func(context.Context) error
		wantErr error
	}{
		{
			name: "hash rejects password over 72 bytes",
			call: func(ctx context.Context) error {
				_, hashErr := hasher.Hash(ctx, tooLongPassword)

				return hashErr
			},
			wantErr: passwordbcrypt.ErrPasswordTooLong,
		},
		{
			name: "compare rejects password over 72 bytes",
			call: func(ctx context.Context) error {
				_, compareErr := hasher.Matches(ctx, "$2a$04$invalid", tooLongPassword)

				return compareErr
			},
			wantErr: passwordbcrypt.ErrPasswordTooLong,
		},
		{
			name: "malformed hash",
			call: func(ctx context.Context) error {
				_, compareErr := hasher.Matches(ctx, "invalid-hash", "password")

				return compareErr
			},
			wantErr: bcryptdriver.ErrHashTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.ErrorIs(t, tt.call(t.Context()), tt.wantErr)
		})
	}
}

func TestHasherContextCancellation(t *testing.T) {
	t.Parallel()

	hasher, err := passwordbcrypt.New(bcryptdriver.MinCost)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err = hasher.Hash(ctx, "password")
	require.ErrorIs(t, err, context.Canceled)

	_, err = hasher.Matches(ctx, "invalid-hash", "password")
	require.ErrorIs(t, err, context.Canceled)
}

func TestNewRejectsInvalidCost(t *testing.T) {
	t.Parallel()

	_, err := passwordbcrypt.New(bcryptdriver.MinCost - 1)
	require.ErrorIs(t, err, passwordbcrypt.ErrInvalidCost)

	_, err = passwordbcrypt.New(bcryptdriver.MaxCost + 1)
	require.ErrorIs(t, err, passwordbcrypt.ErrInvalidCost)
}
