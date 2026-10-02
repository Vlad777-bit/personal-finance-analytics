package grpc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
)

func TestMapServiceError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "canceled", err: context.Canceled, code: codes.Canceled},
		{name: "deadline", err: context.DeadlineExceeded, code: codes.DeadlineExceeded},
		{name: "email required", err: domain.ErrEmailRequired, code: codes.InvalidArgument},
		{name: "invalid email", err: domain.ErrInvalidEmail, code: codes.InvalidArgument},
		{name: "email too long", err: domain.ErrEmailTooLong, code: codes.InvalidArgument},
		{name: "password required", err: domain.ErrPasswordRequired, code: codes.InvalidArgument},
		{name: "password too short", err: domain.ErrPasswordTooShort, code: codes.InvalidArgument},
		{name: "user exists", err: domain.ErrUserAlreadyExists, code: codes.AlreadyExists},
		{name: "invalid credentials", err: domain.ErrInvalidCredentials, code: codes.Unauthenticated},
		{name: "internal", err: errors.New("database error"), code: codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := mapServiceError(fmt.Errorf("service operation: %w", tt.err))
			require.Equal(t, tt.code, status.Code(err))
		})
	}
}
