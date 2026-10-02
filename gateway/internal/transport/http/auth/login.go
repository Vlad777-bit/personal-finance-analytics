package auth

import (
	"net/http"

	authclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/auth"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	request, ok := decodeCredentials(w, r)
	if !ok {
		return
	}

	result, err := h.client.Login(r.Context(), authclient.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		httptransport.WriteAuthError(w, err)

		return
	}

	httptransport.WriteJSON(w, http.StatusOK, loginResponse{
		UserID:      result.UserID,
		Email:       result.Email,
		AccessToken: result.AccessToken,
		ExpiresAt:   result.ExpiresAt,
	})
}
