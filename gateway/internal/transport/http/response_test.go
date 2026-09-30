package httptransport

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
)

func TestWriteLedgerError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "invalid", err: ledgerclient.ErrInvalidArgument, wantStatus: http.StatusBadRequest, wantCode: "invalid_request"},
		{name: "budget", err: ledgerclient.ErrBudgetExceeded, wantStatus: http.StatusConflict, wantCode: "budget_exceeded"},
		{name: "not found", err: ledgerclient.ErrNotFound, wantStatus: http.StatusNotFound, wantCode: "not_found"},
		{name: "canceled", err: ledgerclient.ErrCanceled, wantStatus: http.StatusRequestTimeout, wantCode: "request_canceled"},
		{name: "deadline", err: ledgerclient.ErrDeadline, wantStatus: http.StatusGatewayTimeout, wantCode: "deadline_exceeded"},
		{name: "unavailable", err: ledgerclient.ErrUnavailable, wantStatus: http.StatusServiceUnavailable, wantCode: "service_unavailable"},
		{name: "internal", err: errors.New("unexpected"), wantStatus: http.StatusInternalServerError, wantCode: "internal_error"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			WriteLedgerError(recorder, test.err)

			require.Equal(t, test.wantStatus, recorder.Code)
			require.Contains(t, recorder.Body.String(), `"code":"`+test.wantCode+`"`)
			require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
		})
	}
}
