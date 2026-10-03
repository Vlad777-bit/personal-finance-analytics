package auth

import (
	"errors"
	"time"

	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

type RegisterInput struct {
	Email    string
	Password string
}

type User struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	UserID           string
	Email            string
	AccessToken      string
	ExpiresAt        time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type RefreshResult struct {
	AccessToken string
	ExpiresAt   time.Time
}

func userFromProto(user *authv1.User) (User, error) {
	if user == nil || user.GetId() == "" || user.GetEmail() == "" {
		return User{}, ErrInvalidResponse
	}
	if user.GetCreatedAt() == nil {
		return User{}, errors.Join(
			ErrInvalidResponse,
			errors.New("user created_at is required"),
		)
	}
	if err := user.GetCreatedAt().CheckValid(); err != nil {
		return User{}, errors.Join(ErrInvalidResponse, err)
	}

	return User{
		ID:        user.GetId(),
		Email:     user.GetEmail(),
		CreatedAt: user.GetCreatedAt().AsTime(),
	}, nil
}
