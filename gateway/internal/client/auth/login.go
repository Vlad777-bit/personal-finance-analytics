package auth

import (
	"context"
	"errors"

	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

func (c *Client) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	response, err := c.service.Login(ctx, &authv1.LoginRequest{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		return LoginResult{}, mapError(err)
	}
	if response == nil || response.GetUserId() == "" ||
		response.GetEmail() == "" || response.GetAccessToken() == "" || response.GetRefreshToken() == "" {
		return LoginResult{}, ErrInvalidResponse
	}
	if response.GetExpiresAt() == nil {
		return LoginResult{}, errors.Join(
			ErrInvalidResponse,
			errors.New("access token expires_at is required"),
		)
	}
	if err := response.GetExpiresAt().CheckValid(); err != nil {
		return LoginResult{}, errors.Join(ErrInvalidResponse, err)
	}
	if response.GetRefreshExpiresAt() == nil || response.GetRefreshExpiresAt().CheckValid() != nil {
		return LoginResult{}, ErrInvalidResponse
	}

	return LoginResult{
		UserID:           response.GetUserId(),
		Email:            response.GetEmail(),
		AccessToken:      response.GetAccessToken(),
		ExpiresAt:        response.GetExpiresAt().AsTime(),
		RefreshToken:     response.GetRefreshToken(),
		RefreshExpiresAt: response.GetRefreshExpiresAt().AsTime(),
	}, nil
}
