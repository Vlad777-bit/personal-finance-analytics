package report

import (
	"context"
	"net/http"
	"strings"
	"time"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
	"github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/middleware"
)

type LedgerClient interface {
	GetSummary(
		ctx context.Context,
		input ledgerclient.GetSummaryInput,
	) (ledgerclient.Summary, error)
}

type Handler struct {
	ledgerClient LedgerClient
}

type summaryResponse struct {
	UserID     string                    `json:"user_id"`
	From       time.Time                 `json:"from"`
	To         time.Time                 `json:"to"`
	TotalSpent int64                     `json:"total_spent"`
	Categories []categorySummaryResponse `json:"categories"`
}

type categorySummaryResponse struct {
	Category         string `json:"category"`
	Spent            int64  `json:"spent"`
	BudgetLimit      int64  `json:"budget_limit"`
	BudgetConfigured bool   `json:"budget_configured"`
	Remaining        int64  `json:"remaining"`
	BudgetExceeded   bool   `json:"budget_exceeded"`
}

func NewHandler(client LedgerClient) *Handler {
	return &Handler{ledgerClient: client}
}

func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		httptransport.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")

		return
	}

	query := r.URL.Query()
	fromValue := strings.TrimSpace(query.Get("from"))
	toValue := strings.TrimSpace(query.Get("to"))
	if fromValue == "" || toValue == "" {
		httptransport.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"from and to are required",
		)

		return
	}

	from, err := time.Parse(time.RFC3339, fromValue)
	if err != nil {
		httptransport.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"from must use RFC3339 format",
		)

		return
	}

	to, err := time.Parse(time.RFC3339, toValue)
	if err != nil {
		httptransport.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"to must use RFC3339 format",
		)

		return
	}

	if !from.Before(to) {
		httptransport.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"from must be before to",
		)

		return
	}

	summary, err := h.ledgerClient.GetSummary(
		r.Context(),
		ledgerclient.GetSummaryInput{UserID: identity.UserID, From: from, To: to},
	)
	if err != nil {
		httptransport.WriteLedgerError(w, err)

		return
	}

	categories := make([]categorySummaryResponse, 0, len(summary.Categories))
	for _, category := range summary.Categories {
		categories = append(categories, categorySummaryResponse{
			Category:         category.Category,
			Spent:            category.Spent,
			BudgetLimit:      category.BudgetLimit,
			BudgetConfigured: category.BudgetConfigured,
			Remaining:        category.Remaining,
			BudgetExceeded:   category.BudgetExceeded,
		})
	}

	httptransport.WriteJSON(w, http.StatusOK, summaryResponse{
		UserID:     summary.UserID,
		From:       summary.From,
		To:         summary.To,
		TotalSpent: summary.TotalSpent,
		Categories: categories,
	})
}
