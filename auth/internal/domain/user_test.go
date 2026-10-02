package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()
	maximumLengthEmail := strings.Repeat("a", 64) + "@" +
		strings.Repeat("b", 63) + "." +
		strings.Repeat("c", 63) + "." +
		strings.Repeat("d", 61)

	tests := []struct {
		name    string
		email   string
		want    string
		wantErr error
	}{
		{name: "normalizes email", email: "  User@Example.COM ", want: "user@example.com"},
		{
			name:  "email at maximum length",
			email: maximumLengthEmail,
			want:  maximumLengthEmail,
		},
		{name: "empty email", wantErr: domain.ErrEmailRequired},
		{name: "invalid email", email: "not-an-email", wantErr: domain.ErrInvalidEmail},
		{name: "display name is rejected", email: "User <user@example.com>", wantErr: domain.ErrInvalidEmail},
		{
			name:    "email exceeds maximum length",
			email:   maximumLengthEmail + "x",
			wantErr: domain.ErrEmailTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := domain.NormalizeEmail(tt.email)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if got != tt.want {
				t.Fatalf("expected email %q, got %q", tt.want, got)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{name: "valid password", password: "password"},
		{name: "valid unicode password", password: "пароль12"},
		{name: "empty password", wantErr: domain.ErrPasswordRequired},
		{name: "short password", password: "short", wantErr: domain.ErrPasswordTooShort},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := domain.ValidatePassword(tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestNewUser(t *testing.T) {
	t.Parallel()

	user, err := domain.NewUser(" User@Example.COM ", "password-hash")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if user.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", user.Email)
	}

	_, err = domain.NewUser("user@example.com", "")
	if !errors.Is(err, domain.ErrPasswordHashEmpty) {
		t.Fatalf("expected %v, got %v", domain.ErrPasswordHashEmpty, err)
	}
}
