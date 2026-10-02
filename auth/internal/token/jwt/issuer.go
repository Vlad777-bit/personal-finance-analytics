package jwt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	jwtlibrary "github.com/golang-jwt/jwt/v5"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/service"
)

var (
	ErrSecretRequired = errors.New("JWT secret is required")
	ErrIssuerRequired = errors.New("JWT issuer is required")
	ErrInvalidTTL     = errors.New("JWT TTL must be positive")
)

type claims struct {
	Email string `json:"email"`
	jwtlibrary.RegisteredClaims
}

var _ service.TokenIssuer = (*Issuer)(nil)

type Issuer struct {
	secret []byte
	issuer string
	ttl    time.Duration
	now    func() time.Time
}

func New(secret, issuer string, ttl time.Duration) (*Issuer, error) {
	return newIssuer(secret, issuer, ttl, time.Now)
}

func newIssuer(
	secret string,
	issuer string,
	ttl time.Duration,
	now func() time.Time,
) (*Issuer, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, ErrSecretRequired
	}
	if strings.TrimSpace(issuer) == "" {
		return nil, ErrIssuerRequired
	}
	if ttl <= 0 {
		return nil, ErrInvalidTTL
	}

	return &Issuer{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
		now:    now,
	}, nil
}

func (i *Issuer) Issue(ctx context.Context, user domain.User) (service.AccessToken, error) {
	if err := ctx.Err(); err != nil {
		return service.AccessToken{}, fmt.Errorf("issue JWT: %w", err)
	}

	issuedAt := i.now()
	expiresAt := issuedAt.Add(i.ttl)
	token := jwtlibrary.NewWithClaims(jwtlibrary.SigningMethodHS256, claims{
		Email: user.Email,
		RegisteredClaims: jwtlibrary.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   user.ID,
			IssuedAt:  jwtlibrary.NewNumericDate(issuedAt),
			ExpiresAt: jwtlibrary.NewNumericDate(expiresAt),
		},
	})

	value, err := token.SignedString(i.secret)
	if err != nil {
		return service.AccessToken{}, fmt.Errorf("sign JWT: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return service.AccessToken{}, fmt.Errorf("issue JWT: %w", err)
	}

	return service.AccessToken{Value: value, ExpiresAt: expiresAt}, nil
}
