package transaction

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
	"github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/middleware"
)

const maxCSVImportSize = 10 << 20

type CSVClient interface {
	ImportTransactions(
		ctx context.Context,
		input ledgerclient.ImportTransactionsInput,
	) (ledgerclient.ImportTransactionsResult, error)

	ExportTransactions(
		ctx context.Context,
		input ledgerclient.ExportTransactionsInput,
	) (string, error)
}

type csvHandler struct {
	client CSVClient
}

func NewCSVHandler(client CSVClient) *csvHandler {
	return &csvHandler{client: client}
}

func (h *csvHandler) Import(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		httptransport.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")

		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxCSVImportSize))
	if err != nil {
		httptransport.WriteError(w, http.StatusBadRequest, "invalid_request", "CSV body is too large or unreadable")

		return
	}

	result, err := h.client.ImportTransactions(
		r.Context(),
		ledgerclient.ImportTransactionsInput{
			UserID:  identity.UserID,
			CSVData: string(body),
		},
	)
	if err != nil {
		httptransport.WriteLedgerError(w, err)

		return
	}

	errors := make([]importErrorResponse, 0, len(result.Errors))
	for _, importError := range result.Errors {
		errors = append(errors, importErrorResponse{
			Row:     int64(importError.Row),
			Message: importError.Message,
		})
	}
	httptransport.WriteJSON(w, http.StatusOK, importResponse{
		ImportedCount: result.ImportedCount,
		FailedCount:   result.FailedCount,
		Errors:        errors,
	})
}

type importResponse struct {
	ImportedCount int                   `json:"imported_count"`
	FailedCount   int                   `json:"failed_count"`
	Errors        []importErrorResponse `json:"errors"`
}

type importErrorResponse struct {
	Row     int64  `json:"row"`
	Message string `json:"message"`
}

func (h *csvHandler) Export(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		httptransport.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")

		return
	}

	query := r.URL.Query()
	from, to, ok := parseCSVPeriod(query.Get("from"), query.Get("to"))
	if !ok {
		httptransport.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"from and to must use RFC3339 format and from must be before to",
		)

		return
	}

	csvData, err := h.client.ExportTransactions(
		r.Context(),
		ledgerclient.ExportTransactionsInput{
			UserID:   identity.UserID,
			Category: strings.TrimSpace(query.Get("category")),
			From:     from,
			To:       to,
		},
	)
	if err != nil {
		httptransport.WriteLedgerError(w, err)

		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="transactions.csv"`)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(csvData)); err != nil {
		return
	}
}

func parseCSVPeriod(fromValue, toValue string) (time.Time, time.Time, bool) {
	from, fromErr := time.Parse(time.RFC3339, strings.TrimSpace(fromValue))
	to, toErr := time.Parse(time.RFC3339, strings.TrimSpace(toValue))

	return from, to, fromErr == nil && toErr == nil && from.Before(to)
}
