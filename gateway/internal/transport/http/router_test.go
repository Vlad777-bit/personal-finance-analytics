package httptransport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPing(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	NewRouter().ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
}

func TestOpenAPISpecification(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	recorder := httptest.NewRecorder()

	NewRouter().ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "application/yaml; charset=utf-8", recorder.Header().Get("Content-Type"))
	require.Contains(t, recorder.Body.String(), "openapi: 3.0.3")
	require.Contains(t, recorder.Body.String(), "/transactions/import:")
	require.Contains(t, recorder.Body.String(), "/auth/refresh:")
	require.Contains(t, recorder.Body.String(), "/auth/logout:")
	require.Contains(t, recorder.Body.String(), "refresh_expires_at")
}
