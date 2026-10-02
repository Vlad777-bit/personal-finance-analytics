package jwt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	jwtlibrary "github.com/golang-jwt/jwt/v5"

	"github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/token"
)

var (
	ErrSecretRequired = errors.New("JWT secret is required")
	ErrIssuerRequired = errors.New("JWT issuer is required")
	ErrInvalidToken   = errors.New("JWT is invalid")
)

type claims struct {
	Email string `json:"email"`
	jwtlibrary.RegisteredClaims
}

var _ token.Verifier = (*Verifier)(nil)

type Verifier struct {
	secret []byte
	issuer string
	now    func() time.Time
}

func New(secret, issuer string) (*Verifier, error) {
	return newVerifier(secret, issuer, time.Now)
}

func newVerifier(secret, issuer string, now func() time.Time) (*Verifier, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, ErrSecretRequired
	}
	if strings.TrimSpace(issuer) == "" {
		return nil, ErrIssuerRequired
	}

	return &Verifier{secret: []byte(secret), issuer: issuer, now: now}, nil
}

func (v *Verifier) Verify(ctx context.Context, value string) (token.Identity, error) {
	if err := ctx.Err(); err != nil {
		return token.Identity{}, fmt.Errorf("verify JWT context: %w", err)
	}

	parsedClaims := &claims{}
	parsedToken, err := jwtlibrary.ParseWithClaims(
		value,
		parsedClaims,
		func(*jwtlibrary.Token) (any, error) { return v.secret, nil },
		jwtlibrary.WithValidMethods([]string{jwtlibrary.SigningMethodHS256.Alg()}),
		jwtlibrary.WithIssuer(v.issuer),
		jwtlibrary.WithExpirationRequired(),
		jwtlibrary.WithTimeFunc(v.now),
	)
	if err != nil || !parsedToken.Valid ||
		strings.TrimSpace(parsedClaims.Subject) == "" ||
		strings.TrimSpace(parsedClaims.Email) == "" {
		return token.Identity{}, ErrInvalidToken
	}
	if err := ctx.Err(); err != nil {
		return token.Identity{}, fmt.Errorf("verify JWT context: %w", err)
	}

	return token.Identity{
		UserID: parsedClaims.Subject,
		Email:  parsedClaims.Email,
	}, nil
}
