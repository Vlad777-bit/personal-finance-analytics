package httptransport

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
)

type errorBody struct {
	Error errorDetails `json:"error"`
}

type errorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, value any) {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(value); err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(body.Bytes()); err != nil {
		return
	}
}

func WriteError(
	w http.ResponseWriter,
	statusCode int,
	code string,
	message string,
) {
	WriteJSON(w, statusCode, errorBody{
		Error: errorDetails{
			Code:    code,
			Message: message,
		},
	})
}

func WriteLedgerError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ledgerclient.ErrInvalidArgument):
		WriteError(w, http.StatusBadRequest, "invalid_request", "invalid request")
	case errors.Is(err, ledgerclient.ErrBudgetExceeded):
		WriteError(w, http.StatusConflict, "budget_exceeded", "budget exceeded")
	case errors.Is(err, ledgerclient.ErrNotFound):
		WriteError(w, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, ledgerclient.ErrCanceled):
		WriteError(w, http.StatusRequestTimeout, "request_canceled", "request canceled")
	case errors.Is(err, ledgerclient.ErrDeadline):
		WriteError(w, http.StatusGatewayTimeout, "deadline_exceeded", "upstream timeout")
	case errors.Is(err, ledgerclient.ErrUnavailable):
		WriteError(w, http.StatusServiceUnavailable, "service_unavailable", "service unavailable")
	default:
		WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			http.StatusText(http.StatusInternalServerError),
		)
	}
}
