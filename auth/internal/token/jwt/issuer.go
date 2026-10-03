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
	ErrSecretRequired      = errors.New("JWT secret is required")
	ErrIssuerRequired      = errors.New("JWT issuer is required")
	ErrInvalidTTL          = errors.New("JWT TTL must be positive")
	ErrInvalidRefreshToken = domain.ErrInvalidRefreshToken
)

type claims struct {
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	jwtlibrary.RegisteredClaims
}

var _ service.TokenIssuer = (*Issuer)(nil)

type Issuer struct {
	secret     []byte
	issuer     string
	ttl        time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

func New(secret, issuer string, ttl, refreshTTL time.Duration) (*Issuer, error) {
	return newIssuer(secret, issuer, ttl, refreshTTL, time.Now)
}

func newIssuer(
	secret string,
	issuer string,
	ttl time.Duration,
	refreshTTL time.Duration,
	now func() time.Time,
) (*Issuer, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, ErrSecretRequired
	}
	if strings.TrimSpace(issuer) == "" {
		return nil, ErrIssuerRequired
	}
	if ttl <= 0 || refreshTTL <= 0 {
		return nil, ErrInvalidTTL
	}

	return &Issuer{
		secret:     []byte(secret),
		issuer:     issuer,
		ttl:        ttl,
		refreshTTL: refreshTTL,
		now:        now,
	}, nil
}

func (i *Issuer) Issue(ctx context.Context, user domain.User) (service.AccessToken, error) {
	value, expiresAt, err := i.issue(ctx, user, i.ttl, "access")
	if err != nil {
		return service.AccessToken{}, err
	}
	return service.AccessToken{Value: value, ExpiresAt: expiresAt}, nil
}

func (i *Issuer) IssueRefresh(ctx context.Context, user domain.User) (service.RefreshToken, error) {
	value, expiresAt, err := i.issue(ctx, user, i.refreshTTL, "refresh")
	if err != nil {
		return service.RefreshToken{}, err
	}
	return service.RefreshToken{Value: value, ExpiresAt: expiresAt}, nil
}

func (i *Issuer) Refresh(ctx context.Context, value string) (service.AccessToken, error) {
	pair, err := i.RefreshTokens(ctx, value)
	if err != nil {
		return service.AccessToken{}, err
	}
	return pair.AccessToken, nil
}

func (i *Issuer) RefreshTokens(ctx context.Context, value string) (service.TokenPair, error) {
	if err := ctx.Err(); err != nil {
		return service.TokenPair{}, fmt.Errorf("refresh JWT: %w", err)
	}
	parsedClaims := &claims{}
	_, err := jwtlibrary.ParseWithClaims(value, parsedClaims,
		func(*jwtlibrary.Token) (any, error) { return i.secret, nil },
		jwtlibrary.WithValidMethods([]string{jwtlibrary.SigningMethodHS256.Alg()}),
		jwtlibrary.WithIssuer(i.issuer),
		jwtlibrary.WithTimeFunc(i.now),
	)
	if err != nil || parsedClaims.TokenType != "refresh" || parsedClaims.Subject == "" || parsedClaims.Email == "" {
		return service.TokenPair{}, ErrInvalidRefreshToken
	}
	user := domain.User{ID: parsedClaims.Subject, Email: parsedClaims.Email}
	accessValue, accessExpiresAt, err := i.issue(ctx, user, i.ttl, "access")
	if err != nil {
		return service.TokenPair{}, err
	}
	refreshValue, refreshExpiresAt, err := i.issue(ctx, user, i.refreshTTL, "refresh")
	if err != nil {
		return service.TokenPair{}, err
	}
	return service.TokenPair{
		UserID:       parsedClaims.Subject,
		AccessToken:  service.AccessToken{Value: accessValue, ExpiresAt: accessExpiresAt},
		RefreshToken: service.RefreshToken{Value: refreshValue, ExpiresAt: refreshExpiresAt},
	}, nil
}

func (i *Issuer) issue(ctx context.Context, user domain.User, ttl time.Duration, tokenType string) (string, time.Time, error) {
	if err := ctx.Err(); err != nil {
		return "", time.Time{}, fmt.Errorf("issue JWT: %w", err)
	}
	issuedAt := i.now()
	expiresAt := issuedAt.Add(ttl)
	token := jwtlibrary.NewWithClaims(jwtlibrary.SigningMethodHS256, claims{Email: user.Email, TokenType: tokenType, RegisteredClaims: jwtlibrary.RegisteredClaims{Issuer: i.issuer, Subject: user.ID, IssuedAt: jwtlibrary.NewNumericDate(issuedAt), ExpiresAt: jwtlibrary.NewNumericDate(expiresAt)}})
	value, err := token.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign JWT: %w", err)
	}
	return value, expiresAt, nil
}
