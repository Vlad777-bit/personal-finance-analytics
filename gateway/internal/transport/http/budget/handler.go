package budget

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
)

const maxRequestBodySize = 1 << 20

type LedgerClient interface {
	CreateBudget(
		ctx context.Context,
		input ledgerclient.CreateBudgetInput,
	) (ledgerclient.Budget, error)
}

type Handler struct {
	ledgerClient LedgerClient
}

type upsertRequest struct {
	UserID string `json:"user_id"`
	Limit  int64  `json:"limit_amount"`
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
	request, err := decodeUpsertRequest(w, r)
	if err != nil {
		httptransport.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())

		return
	}

	category := strings.TrimSpace(r.PathValue("category"))
	if category == "" || strings.TrimSpace(request.UserID) == "" || request.Limit <= 0 {
		httptransport.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"user_id, category and positive limit_amount are required",
		)

		return
	}

	createdBudget, err := h.ledgerClient.CreateBudget(
		r.Context(),
		ledgerclient.CreateBudgetInput{
			UserID:   request.UserID,
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

func decodeUpsertRequest(
	w http.ResponseWriter,
	r *http.Request,
) (upsertRequest, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request upsertRequest
	if err := decoder.Decode(&request); err != nil {
		return upsertRequest{}, errors.New("request body must contain valid JSON")
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return upsertRequest{}, errors.New("request body must contain a single JSON object")
	}

	return request, nil
}
