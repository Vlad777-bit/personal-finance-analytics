package auth

import (
	"context"
	"errors"

	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

func (c *Client) Refresh(ctx context.Context, refreshToken string) (RefreshResult, error) {
	if refreshToken == "" {
		return RefreshResult{}, ErrInvalidArgument
	}
	response, err := c.service.Refresh(ctx, &authv1.RefreshRequest{RefreshToken: refreshToken})
	if err != nil {
		return RefreshResult{}, mapError(err)
	}
	if response == nil || response.GetAccessToken() == "" || response.GetRefreshToken() == "" || response.GetExpiresAt() == nil || response.GetRefreshExpiresAt() == nil {
		return RefreshResult{}, ErrInvalidResponse
	}
	if err := response.GetExpiresAt().CheckValid(); err != nil {
		return RefreshResult{}, errors.Join(ErrInvalidResponse, err)
	}
	if err := response.GetRefreshExpiresAt().CheckValid(); err != nil {
		return RefreshResult{}, errors.Join(ErrInvalidResponse, err)
	}
	return RefreshResult{AccessToken: response.GetAccessToken(), ExpiresAt: response.GetExpiresAt().AsTime(), RefreshToken: response.GetRefreshToken(), RefreshExpiresAt: response.GetRefreshExpiresAt().AsTime()}, nil
}
