package auth

import (
	"net/http"
	"strings"

	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := httptransport.DecodeJSON(w, r, &request); err != nil {
		httptransport.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if strings.TrimSpace(request.RefreshToken) == "" {
		httptransport.WriteError(w, http.StatusBadRequest, "invalid_request", "refresh_token is required")
		return
	}
	if err := h.client.Logout(r.Context(), request.RefreshToken); err != nil {
		httptransport.WriteAuthError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
