package auth_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	authclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/auth"
	authtransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/auth"
)

type fakeAuthClient struct {
	register func(context.Context, authclient.RegisterInput) (authclient.User, error)
	login    func(context.Context, authclient.LoginInput) (authclient.LoginResult, error)
	refresh  func(context.Context, string) (authclient.RefreshResult, error)
	logout   func(context.Context, string) error
}

func (f *fakeAuthClient) Refresh(ctx context.Context, token string) (authclient.RefreshResult, error) {
	if f.refresh == nil {
		return authclient.RefreshResult{}, errors.New("unexpected Refresh call")
	}
	return f.refresh(ctx, token)
}

func (f *fakeAuthClient) Logout(ctx context.Context, token string) error {
	if f.logout == nil {
		return errors.New("unexpected Logout call")
	}
	return f.logout(ctx, token)
}

func (f *fakeAuthClient) Register(
	ctx context.Context,
	input authclient.RegisterInput,
) (authclient.User, error) {
	return f.register(ctx, input)
}

func (f *fakeAuthClient) Login(
	ctx context.Context,
	input authclient.LoginInput,
) (authclient.LoginResult, error) {
	return f.login(ctx, input)
}

func TestHandler_Register(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		body       string
		client     *fakeAuthClient
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			body: `{"email":"user@example.com","password":"secure-password"}`,
			client: &fakeAuthClient{register: func(
				_ context.Context,
				input authclient.RegisterInput,
			) (authclient.User, error) {
				require.Equal(t, "user@example.com", input.Email)
				require.Equal(t, "secure-password", input.Password)

				return authclient.User{
					ID: "user-1", Email: "user@example.com", CreatedAt: createdAt,
				}, nil
			}},
			wantStatus: http.StatusCreated,
			wantBody: `{
				"id":"user-1","email":"user@example.com",
				"created_at":"2026-10-02T12:00:00Z"
			}`,
		},
		{
			name:       "malformed JSON",
			body:       `{`,
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_request","message":"request body must contain valid JSON"}}`,
		},
		{
			name:       "missing credentials",
			body:       `{"email":"","password":""}`,
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_request","message":"email and password are required"}}`,
		},
		{
			name: "duplicate user",
			body: `{"email":"user@example.com","password":"secure-password"}`,
			client: &fakeAuthClient{register: func(
				context.Context,
				authclient.RegisterInput,
			) (authclient.User, error) {
				return authclient.User{}, authclient.ErrAlreadyExists
			}},
			wantStatus: http.StatusConflict,
			wantBody:   `{"error":{"code":"user_already_exists","message":"user already exists"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()
			authtransport.NewHandler(tt.client).Register(recorder, request)

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.JSONEq(t, tt.wantBody, recorder.Body.String())
		})
	}
}

func TestHandler_Login(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, time.October, 2, 12, 15, 0, 0, time.UTC)
	tests := []struct {
		name       string
		body       string
		client     *fakeAuthClient
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			body: `{"email":"user@example.com","password":"secure-password"}`,
			client: &fakeAuthClient{login: func(
				_ context.Context,
				input authclient.LoginInput,
			) (authclient.LoginResult, error) {
				require.Equal(t, "user@example.com", input.Email)
				require.Equal(t, "secure-password", input.Password)

				return authclient.LoginResult{
					UserID: "user-1", Email: "user@example.com",
					AccessToken: "access-token", ExpiresAt: expiresAt,
					RefreshToken: "refresh-token", RefreshExpiresAt: expiresAt.Add(24 * time.Hour),
				}, nil
			}},
			wantStatus: http.StatusOK,
			wantBody: `{
				"user_id":"user-1","email":"user@example.com",
				"access_token":"access-token","expires_at":"2026-10-02T12:15:00Z",
				"refresh_token":"refresh-token","refresh_expires_at":"2026-10-03T12:15:00Z"
			}`,
		},
		{
			name: "invalid credentials",
			body: `{"email":"user@example.com","password":"wrong-password"}`,
			client: &fakeAuthClient{login: func(
				context.Context,
				authclient.LoginInput,
			) (authclient.LoginResult, error) {
				return authclient.LoginResult{}, authclient.ErrInvalidCredentials
			}},
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"error":{"code":"invalid_credentials","message":"invalid credentials"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()
			authtransport.NewHandler(tt.client).Login(recorder, request)

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.JSONEq(t, tt.wantBody, recorder.Body.String())
		})
	}
}

func TestHandler_Refresh(t *testing.T) {
	t.Parallel()
	expiresAt := time.Date(2026, time.October, 2, 12, 15, 0, 0, time.UTC)
	client := &fakeAuthClient{refresh: func(_ context.Context, token string) (authclient.RefreshResult, error) {
		require.Equal(t, "refresh-token", token)
		return authclient.RefreshResult{AccessToken: "new-access", ExpiresAt: expiresAt, RefreshToken: "new-refresh", RefreshExpiresAt: expiresAt.Add(24 * time.Hour)}, nil
	}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(`{"refresh_token":"refresh-token"}`))
	authtransport.NewHandler(client).Refresh(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"access_token":"new-access","expires_at":"2026-10-02T12:15:00Z","refresh_token":"new-refresh","refresh_expires_at":"2026-10-03T12:15:00Z"}`, recorder.Body.String())
}

func TestHandler_Logout(t *testing.T) {
	t.Parallel()
	client := &fakeAuthClient{logout: func(_ context.Context, token string) error {
		require.Equal(t, "refresh-token", token)
		return nil
	}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader(`{"refresh_token":"refresh-token"}`))
	authtransport.NewHandler(client).Logout(recorder, request)
	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func clientMustNotBeCalled(t *testing.T) *fakeAuthClient {
	t.Helper()

	return &fakeAuthClient{
		register: func(context.Context, authclient.RegisterInput) (authclient.User, error) {
			return authclient.User{}, errors.New("unexpected Register call")
		},
		login: func(context.Context, authclient.LoginInput) (authclient.LoginResult, error) {
			return authclient.LoginResult{}, errors.New("unexpected Login call")
		},
		refresh: func(context.Context, string) (authclient.RefreshResult, error) {
			return authclient.RefreshResult{}, errors.New("unexpected Refresh call")
		},
	}
}
