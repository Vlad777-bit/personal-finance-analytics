package auth

import (
	"net/http"

	authclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/auth"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	request, ok := decodeCredentials(w, r)
	if !ok {
		return
	}

	user, err := h.client.Register(r.Context(), authclient.RegisterInput{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		httptransport.WriteAuthError(w, err)

		return
	}

	httptransport.WriteJSON(w, http.StatusCreated, userResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}
