package auth

import (
	"net/http"
	"strings"

	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
)

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := httptransport.DecodeJSON(w, r, &request); err != nil {
		httptransport.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if strings.TrimSpace(request.RefreshToken) == "" {
		httptransport.WriteError(w, http.StatusBadRequest, "invalid_request", "refresh_token is required")
		return
	}
	result, err := h.client.Refresh(r.Context(), request.RefreshToken)
	if err != nil {
		httptransport.WriteAuthError(w, err)
		return
	}
	httptransport.WriteJSON(w, http.StatusOK, refreshResponse{AccessToken: result.AccessToken, ExpiresAt: result.ExpiresAt})
}
