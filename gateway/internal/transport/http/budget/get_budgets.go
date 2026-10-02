package budget

import (
	"net/http"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
	"github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/middleware"
)

func (h *Handler) GetBudgets(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		httptransport.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")

		return
	}

	budgets, err := h.ledgerClient.GetBudgets(
		r.Context(),
		ledgerclient.GetBudgetsInput{UserID: identity.UserID},
	)
	if err != nil {
		httptransport.WriteLedgerError(w, err)

		return
	}

	response := make([]budgetResponse, 0, len(budgets))
	for _, budget := range budgets {
		response = append(response, budgetResponse{
			ID:       budget.ID,
			UserID:   budget.UserID,
			Category: budget.Category,
			Limit:    budget.Limit,
		})
	}

	httptransport.WriteJSON(w, http.StatusOK, response)
}
