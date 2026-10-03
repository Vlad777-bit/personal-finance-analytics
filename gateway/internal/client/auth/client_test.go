package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

type fakeAuthServiceClient struct {
	register func(context.Context, *authv1.RegisterRequest) (*authv1.RegisterResponse, error)
	login    func(context.Context, *authv1.LoginRequest) (*authv1.LoginResponse, error)
	refresh  func(context.Context, *authv1.RefreshRequest) (*authv1.RefreshResponse, error)
}

func (f *fakeAuthServiceClient) Refresh(ctx context.Context, request *authv1.RefreshRequest, _ ...grpc.CallOption) (*authv1.RefreshResponse, error) {
	return f.refresh(ctx, request)
}

func (f *fakeAuthServiceClient) Register(
	ctx context.Context,
	request *authv1.RegisterRequest,
	_ ...grpc.CallOption,
) (*authv1.RegisterResponse, error) {
	return f.register(ctx, request)
}

func (f *fakeAuthServiceClient) Login(
	ctx context.Context,
	request *authv1.LoginRequest,
	_ ...grpc.CallOption,
) (*authv1.LoginResponse, error) {
	return f.login(ctx, request)
}

func TestClient_Register(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		call    func(*testing.T, *authv1.RegisterRequest) (*authv1.RegisterResponse, error)
		want    User
		wantErr error
	}{
		{
			name: "success",
			call: func(t *testing.T, request *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
				t.Helper()
				require.Equal(t, "user@example.com", request.GetEmail())
				require.Equal(t, "password", request.GetPassword())

				return &authv1.RegisterResponse{User: &authv1.User{
					Id: "user-1", Email: "user@example.com",
					CreatedAt: timestamppb.New(createdAt),
				}}, nil
			},
			want: User{ID: "user-1", Email: "user@example.com", CreatedAt: createdAt},
		},
		{
			name: "grpc error",
			call: func(*testing.T, *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
				return nil, status.Error(codes.AlreadyExists, "user exists")
			},
			wantErr: ErrAlreadyExists,
		},
		{
			name: "nil response",
			call: func(*testing.T, *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
				return nil, nil
			},
			wantErr: ErrInvalidResponse,
		},
		{
			name: "missing user",
			call: func(*testing.T, *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
				return &authv1.RegisterResponse{}, nil
			},
			wantErr: ErrInvalidResponse,
		},
		{
			name: "invalid timestamp",
			call: func(*testing.T, *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
				return &authv1.RegisterResponse{User: &authv1.User{
					Id: "user-1", Email: "user@example.com",
					CreatedAt: &timestamppb.Timestamp{Seconds: 253402300800},
				}}, nil
			},
			wantErr: ErrInvalidResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &Client{service: &fakeAuthServiceClient{register: func(
				_ context.Context,
				request *authv1.RegisterRequest,
			) (*authv1.RegisterResponse, error) {
				return tt.call(t, request)
			}}}

			got, err := client.Register(t.Context(), RegisterInput{
				Email: "user@example.com", Password: "password",
			})
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestClient_Login(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, time.October, 2, 12, 15, 0, 0, time.UTC)
	refreshExpiresAt := expiresAt.Add(24 * time.Hour)
	tests := []struct {
		name    string
		call    func(*testing.T, *authv1.LoginRequest) (*authv1.LoginResponse, error)
		want    LoginResult
		wantErr error
	}{
		{
			name: "success",
			call: func(t *testing.T, request *authv1.LoginRequest) (*authv1.LoginResponse, error) {
				t.Helper()
				require.Equal(t, "user@example.com", request.GetEmail())
				require.Equal(t, "password", request.GetPassword())

				return &authv1.LoginResponse{
					UserId: "user-1", Email: "user@example.com",
					AccessToken: "access-token", ExpiresAt: timestamppb.New(expiresAt),
					RefreshToken: "refresh-token", RefreshExpiresAt: timestamppb.New(refreshExpiresAt),
				}, nil
			},
			want: LoginResult{
				UserID: "user-1", Email: "user@example.com",
				AccessToken: "access-token", ExpiresAt: expiresAt,
				RefreshToken: "refresh-token", RefreshExpiresAt: refreshExpiresAt,
			},
		},
		{
			name: "grpc error",
			call: func(*testing.T, *authv1.LoginRequest) (*authv1.LoginResponse, error) {
				return nil, status.Error(codes.Unauthenticated, "invalid credentials")
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "nil response",
			call: func(*testing.T, *authv1.LoginRequest) (*authv1.LoginResponse, error) {
				return nil, nil
			},
			wantErr: ErrInvalidResponse,
		},
		{
			name: "missing token",
			call: func(*testing.T, *authv1.LoginRequest) (*authv1.LoginResponse, error) {
				return &authv1.LoginResponse{UserId: "user-1", Email: "user@example.com"}, nil
			},
			wantErr: ErrInvalidResponse,
		},
		{
			name: "invalid timestamp",
			call: func(*testing.T, *authv1.LoginRequest) (*authv1.LoginResponse, error) {
				return &authv1.LoginResponse{
					UserId: "user-1", Email: "user@example.com", AccessToken: "token",
					ExpiresAt:    &timestamppb.Timestamp{Seconds: 253402300800},
					RefreshToken: "refresh-token", RefreshExpiresAt: timestamppb.New(refreshExpiresAt),
				}, nil
			},
			wantErr: ErrInvalidResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &Client{service: &fakeAuthServiceClient{login: func(
				_ context.Context,
				request *authv1.LoginRequest,
			) (*authv1.LoginResponse, error) {
				return tt.call(t, request)
			}}}

			got, err := client.Login(t.Context(), LoginInput{
				Email: "user@example.com", Password: "password",
			})
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestMapError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "invalid argument", err: status.Error(codes.InvalidArgument, "invalid"), want: ErrInvalidArgument},
		{name: "already exists", err: status.Error(codes.AlreadyExists, "exists"), want: ErrAlreadyExists},
		{name: "unauthenticated", err: status.Error(codes.Unauthenticated, "invalid"), want: ErrInvalidCredentials},
		{name: "canceled", err: status.Error(codes.Canceled, "canceled"), want: ErrCanceled},
		{name: "deadline", err: status.Error(codes.DeadlineExceeded, "deadline"), want: ErrDeadline},
		{name: "unavailable", err: status.Error(codes.Unavailable, "unavailable"), want: ErrUnavailable},
		{name: "unknown code", err: status.Error(codes.Unknown, "unknown"), want: ErrInternal},
		{name: "plain error", err: errors.New("plain"), want: ErrInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.ErrorIs(t, mapError(tt.err), tt.want)
		})
	}
}
