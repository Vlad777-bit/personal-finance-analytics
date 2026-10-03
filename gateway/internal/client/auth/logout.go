package auth

import (
	"context"

	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

func (c *Client) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return ErrInvalidArgument
	}
	if _, err := c.service.Logout(ctx, &authv1.LogoutRequest{RefreshToken: refreshToken}); err != nil {
		return mapError(err)
	}
	return nil
}
