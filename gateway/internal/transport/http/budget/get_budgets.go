package budget

import (
	"net/http"
	"strings"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
)

func (h *Handler) GetBudgets(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	if userID == "" {
		httptransport.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"user_id is required",
		)

		return
	}

	budgets, err := h.ledgerClient.GetBudgets(
		r.Context(),
		ledgerclient.GetBudgetsInput{UserID: userID},
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
