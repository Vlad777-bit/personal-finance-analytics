package auth

import (
	"context"

	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

func (c *Client) Register(ctx context.Context, input RegisterInput) (User, error) {
	response, err := c.service.Register(ctx, &authv1.RegisterRequest{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		return User{}, mapError(err)
	}
	if response == nil {
		return User{}, ErrInvalidResponse
	}

	return userFromProto(response.GetUser())
}
