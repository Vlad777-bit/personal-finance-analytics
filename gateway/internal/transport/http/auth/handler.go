package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	authclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/auth"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
)

type Client interface {
	Register(ctx context.Context, input authclient.RegisterInput) (authclient.User, error)
	Login(ctx context.Context, input authclient.LoginInput) (authclient.LoginResult, error)
	Refresh(ctx context.Context, refreshToken string) (authclient.RefreshResult, error)
}

type Handler struct {
	client Client
}

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type loginResponse struct {
	UserID           string    `json:"user_id"`
	Email            string    `json:"email"`
	AccessToken      string    `json:"access_token"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshToken     string    `json:"refresh_token"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type (
	refreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}
	refreshResponse struct {
		AccessToken string    `json:"access_token"`
		ExpiresAt   time.Time `json:"expires_at"`
	}
)

func NewHandler(client Client) *Handler {
	return &Handler{client: client}
}

func decodeCredentials(
	w http.ResponseWriter,
	r *http.Request,
) (credentialsRequest, bool) {
	var request credentialsRequest
	if err := httptransport.DecodeJSON(w, r, &request); err != nil {
		httptransport.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())

		return credentialsRequest{}, false
	}
	if strings.TrimSpace(request.Email) == "" || request.Password == "" {
		httptransport.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"email and password are required",
		)

		return credentialsRequest{}, false
	}

	return request, true
}
