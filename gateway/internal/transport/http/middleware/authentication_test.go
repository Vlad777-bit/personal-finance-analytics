package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/token"
	"github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/middleware"
)

type fakeVerifier struct {
	verify func(context.Context, string) (token.Identity, error)
}

func (f *fakeVerifier) Verify(ctx context.Context, value string) (token.Identity, error) {
	return f.verify(ctx, value)
}

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		header     string
		verifier   *fakeVerifier
		wantStatus int
		wantCalled bool
	}{
		{name: "missing header", verifier: verifierMustNotBeCalled(t), wantStatus: http.StatusUnauthorized},
		{name: "wrong scheme", header: "Basic token", verifier: verifierMustNotBeCalled(t), wantStatus: http.StatusUnauthorized},
		{name: "missing token", header: "Bearer", verifier: verifierMustNotBeCalled(t), wantStatus: http.StatusUnauthorized},
		{name: "extra value", header: "Bearer token extra", verifier: verifierMustNotBeCalled(t), wantStatus: http.StatusUnauthorized},
		{
			name:   "invalid token",
			header: "Bearer invalid",
			verifier: &fakeVerifier{verify: func(context.Context, string) (token.Identity, error) {
				return token.Identity{}, errors.New("invalid token")
			}},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "valid token",
			header: "bearer valid-token",
			verifier: &fakeVerifier{verify: func(_ context.Context, value string) (token.Identity, error) {
				require.Equal(t, "valid-token", value)

				return token.Identity{UserID: "user-1", Email: "user@example.com"}, nil
			}},
			wantStatus: http.StatusNoContent,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				identity, ok := middleware.IdentityFromContext(r.Context())
				require.True(t, ok)
				require.Equal(t, "user-1", identity.UserID)
				require.Equal(t, "user@example.com", identity.Email)
				w.WriteHeader(http.StatusNoContent)
			})
			handler := middleware.Authenticate(tt.verifier)(next)
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.Header.Set("Authorization", tt.header)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.Equal(t, tt.wantCalled, called)
			if tt.wantStatus == http.StatusUnauthorized {
				require.JSONEq(t, `{"error":{"code":"unauthorized","message":"valid bearer token is required"}}`, recorder.Body.String())
			}
		})
	}
}

func verifierMustNotBeCalled(t *testing.T) *fakeVerifier {
	t.Helper()

	return &fakeVerifier{verify: func(context.Context, string) (token.Identity, error) {
		return token.Identity{}, errors.New("unexpected Verify call")
	}}
}
