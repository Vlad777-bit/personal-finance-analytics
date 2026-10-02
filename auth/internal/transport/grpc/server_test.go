package grpc_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/service"
	servicemocks "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/service/mocks"
	grpctransport "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/transport/grpc"
	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

func TestServer_Register(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)
	authService := servicemocks.NewAuthService(t)
	authService.EXPECT().Register(
		mock.Anything,
		service.RegisterInput{Email: "user@example.com", Password: "password"},
	).Return(domain.User{
		ID:           "user-1",
		Email:        "user@example.com",
		PasswordHash: "must-not-be-exposed",
		CreatedAt:    createdAt,
	}, nil)

	response, err := grpctransport.New(authService).Register(
		t.Context(),
		&authv1.RegisterRequest{Email: "user@example.com", Password: "password"},
	)
	require.NoError(t, err)
	require.Equal(t, "user-1", response.GetUser().GetId())
	require.Equal(t, "user@example.com", response.GetUser().GetEmail())
	require.True(t, createdAt.Equal(response.GetUser().GetCreatedAt().AsTime()))
}

func TestServer_Login(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, time.October, 2, 12, 15, 0, 0, time.UTC)
	authService := servicemocks.NewAuthService(t)
	authService.EXPECT().Login(
		mock.Anything,
		service.LoginInput{Email: "user@example.com", Password: "password"},
	).Return(service.LoginResult{
		UserID: "user-1",
		Email:  "user@example.com",
		AccessToken: service.AccessToken{
			Value:     "access-token",
			ExpiresAt: expiresAt,
		},
	}, nil)

	response, err := grpctransport.New(authService).Login(
		t.Context(),
		&authv1.LoginRequest{Email: "user@example.com", Password: "password"},
	)
	require.NoError(t, err)
	require.Equal(t, "user-1", response.GetUserId())
	require.Equal(t, "user@example.com", response.GetEmail())
	require.Equal(t, "access-token", response.GetAccessToken())
	require.True(t, expiresAt.Equal(response.GetExpiresAt().AsTime()))
}

func TestServer_RejectsNilRequest(t *testing.T) {
	t.Parallel()

	server := grpctransport.New(servicemocks.NewAuthService(t))

	_, registerErr := server.Register(t.Context(), nil)
	_, loginErr := server.Login(t.Context(), nil)

	require.Equal(t, codes.InvalidArgument, status.Code(registerErr))
	require.Equal(t, codes.InvalidArgument, status.Code(loginErr))
}

func TestServer_MapsServiceErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		method       string
		serviceError error
		wantCode     codes.Code
	}{
		{
			name:         "register duplicate",
			method:       "register",
			serviceError: domain.ErrUserAlreadyExists,
			wantCode:     codes.AlreadyExists,
		},
		{
			name:         "login invalid credentials",
			method:       "login",
			serviceError: domain.ErrInvalidCredentials,
			wantCode:     codes.Unauthenticated,
		},
		{
			name:         "internal error",
			method:       "login",
			serviceError: errors.New("database contains sensitive details"),
			wantCode:     codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			authService := servicemocks.NewAuthService(t)
			server := grpctransport.New(authService)
			var err error

			switch tt.method {
			case "register":
				authService.EXPECT().Register(mock.Anything, mock.Anything).
					Return(domain.User{}, tt.serviceError)
				_, err = server.Register(t.Context(), &authv1.RegisterRequest{})
			case "login":
				authService.EXPECT().Login(mock.Anything, mock.Anything).
					Return(service.LoginResult{}, tt.serviceError)
				_, err = server.Login(t.Context(), &authv1.LoginRequest{})
			default:
				require.FailNow(t, "unsupported method", tt.method)
			}

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode == codes.Internal {
				require.Equal(t, "internal error", status.Convert(err).Message())
			}
		})
	}
}
