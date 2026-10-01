package transaction

import (
	"context"
	"net/http"
	"strings"
	"time"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
)

type LedgerClient interface {
	CreateTransaction(
		ctx context.Context,
		input ledgerclient.CreateTransactionInput,
	) (ledgerclient.Transaction, error)

	GetTransactions(
		ctx context.Context,
		input ledgerclient.GetTransactionsInput,
	) ([]ledgerclient.Transaction, error)
}

type Handler struct {
	ledgerClient LedgerClient
}

type createRequest struct {
	UserID      string `json:"user_id"`
	Amount      int64  `json:"amount"`
	Category    string `json:"category"`
	Description string `json:"description"`
	OccurredAt  string `json:"occurred_at"`
}

type transactionResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Amount      int64     `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	OccurredAt  time.Time `json:"occurred_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewHandler(client LedgerClient) *Handler {
	return &Handler{ledgerClient: client}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request createRequest
	if err := httptransport.DecodeJSON(w, r, &request); err != nil {
		httptransport.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())

		return
	}

	if strings.TrimSpace(request.UserID) == "" ||
		request.Amount <= 0 ||
		strings.TrimSpace(request.Category) == "" ||
		strings.TrimSpace(request.OccurredAt) == "" {
		httptransport.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"user_id, positive amount, category and occurred_at are required",
		)

		return
	}

	occurredAt, err := time.Parse(time.RFC3339, request.OccurredAt)
	if err != nil {
		httptransport.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"occurred_at must use RFC3339 format",
		)

		return
	}

	createdTransaction, err := h.ledgerClient.CreateTransaction(
		r.Context(),
		ledgerclient.CreateTransactionInput{
			UserID:      request.UserID,
			Amount:      request.Amount,
			Category:    request.Category,
			Description: request.Description,
			OccurredAt:  occurredAt,
		},
	)
	if err != nil {
		httptransport.WriteLedgerError(w, err)

		return
	}

	httptransport.WriteJSON(w, http.StatusCreated, transactionResponse{
		ID:          createdTransaction.ID,
		UserID:      createdTransaction.UserID,
		Amount:      createdTransaction.Amount,
		Category:    createdTransaction.Category,
		Description: createdTransaction.Description,
		OccurredAt:  createdTransaction.OccurredAt,
		CreatedAt:   createdTransaction.CreatedAt,
	})
}
