package transaction

import (
	"net/http"
	"strings"
	"time"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
)

func (h *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	userID := strings.TrimSpace(query.Get("user_id"))
	fromValue := strings.TrimSpace(query.Get("from"))
	toValue := strings.TrimSpace(query.Get("to"))
	if userID == "" || fromValue == "" || toValue == "" {
		httptransport.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"user_id, from and to are required",
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

	transactions, err := h.ledgerClient.GetTransactions(
		r.Context(),
		ledgerclient.GetTransactionsInput{
			UserID:   userID,
			Category: strings.TrimSpace(query.Get("category")),
			From:     from,
			To:       to,
		},
	)
	if err != nil {
		httptransport.WriteLedgerError(w, err)

		return
	}

	response := make([]transactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		response = append(response, transactionResponse{
			ID:          transaction.ID,
			UserID:      transaction.UserID,
			Amount:      transaction.Amount,
			Category:    transaction.Category,
			Description: transaction.Description,
			OccurredAt:  transaction.OccurredAt,
			CreatedAt:   transaction.CreatedAt,
		})
	}

	httptransport.WriteJSON(w, http.StatusOK, response)
}
