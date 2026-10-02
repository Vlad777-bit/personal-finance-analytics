package bcrypt

import (
	"context"
	"errors"
	"fmt"

	bcryptdriver "golang.org/x/crypto/bcrypt"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/service"
)

const maximumPasswordBytes = 72

var (
	ErrInvalidCost     = errors.New("bcrypt cost is invalid")
	ErrPasswordTooLong = errors.New("password exceeds bcrypt 72-byte limit")
)

var _ service.PasswordHasher = (*Hasher)(nil)

type Hasher struct {
	cost int
}

func New(cost int) (*Hasher, error) {
	if cost < bcryptdriver.MinCost || cost > bcryptdriver.MaxCost {
		return nil, fmt.Errorf(
			"%w: must be between %d and %d",
			ErrInvalidCost,
			bcryptdriver.MinCost,
			bcryptdriver.MaxCost,
		)
	}

	return &Hasher{cost: cost}, nil
}

func (h *Hasher) Hash(ctx context.Context, password string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("hash password context: %w", err)
	}
	if len(password) > maximumPasswordBytes {
		return "", ErrPasswordTooLong
	}

	passwordHash, err := bcryptdriver.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		if errors.Is(err, bcryptdriver.ErrPasswordTooLong) {
			return "", ErrPasswordTooLong
		}

		return "", fmt.Errorf("generate bcrypt hash: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("hash password context: %w", err)
	}

	return string(passwordHash), nil
}

func (h *Hasher) Matches(
	ctx context.Context,
	passwordHash string,
	password string,
) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, fmt.Errorf("compare password context: %w", err)
	}
	if len(password) > maximumPasswordBytes {
		return false, ErrPasswordTooLong
	}

	err := bcryptdriver.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if errors.Is(err, bcryptdriver.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("compare bcrypt hash: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return false, fmt.Errorf("compare password context: %w", err)
	}

	return true, nil
}
