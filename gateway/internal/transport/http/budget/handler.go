package budget

import (
	"context"
	"net/http"
	"strings"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
	"github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/middleware"
)

type LedgerClient interface {
	CreateBudget(
		ctx context.Context,
		input ledgerclient.CreateBudgetInput,
	) (ledgerclient.Budget, error)

	GetBudgets(
		ctx context.Context,
		input ledgerclient.GetBudgetsInput,
	) ([]ledgerclient.Budget, error)
}

type Handler struct {
	ledgerClient LedgerClient
}

type upsertRequest struct {
	Limit int64 `json:"limit_amount"`
}

type budgetResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Category string `json:"category"`
	Limit    int64  `json:"limit_amount"`
}

func NewHandler(client LedgerClient) *Handler {
	return &Handler{ledgerClient: client}
}

func (h *Handler) Upsert(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		httptransport.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")

		return
	}

	var request upsertRequest
	err := httptransport.DecodeJSON(w, r, &request)
	if err != nil {
		httptransport.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())

		return
	}

	category := strings.TrimSpace(r.PathValue("category"))
	if category == "" || request.Limit <= 0 {
		httptransport.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"category and positive limit_amount are required",
		)

		return
	}

	createdBudget, err := h.ledgerClient.CreateBudget(
		r.Context(),
		ledgerclient.CreateBudgetInput{
			UserID:   identity.UserID,
			Category: category,
			Limit:    request.Limit,
		},
	)
	if err != nil {
		httptransport.WriteLedgerError(w, err)

		return
	}

	httptransport.WriteJSON(w, http.StatusOK, budgetResponse{
		ID:       createdBudget.ID,
		UserID:   createdBudget.UserID,
		Category: createdBudget.Category,
		Limit:    createdBudget.Limit,
	})
}
