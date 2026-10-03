package auth

import (
	"net/http"

	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
	authmiddleware "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/middleware"
)

func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	identity, ok := authmiddleware.IdentityFromContext(r.Context())
	if !ok || identity.UserID == "" {
		httptransport.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	if err := h.client.LogoutAll(r.Context(), identity.UserID); err != nil {
		httptransport.WriteAuthError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
