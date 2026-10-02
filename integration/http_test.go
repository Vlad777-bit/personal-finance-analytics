//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	testCategory = "http-e2e"
	testEmail    = "http-e2e@example.com"
)

type runningProcess struct {
	command *exec.Cmd
	done    chan error
}

type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (t bearerTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	request = request.Clone(request.Context())
	request.Header.Set("Authorization", "Bearer "+t.token)

	return t.base.RoundTrip(request)
}

func TestGatewayLedgerHTTP(t *testing.T) {
	databaseURL := requireEnv(t, "LEDGER_DATABASE_URL")
	repositoryRoot := filepath.Clean("..")
	temporaryDirectory := t.TempDir()

	ledgerBinary := filepath.Join(temporaryDirectory, "ledger")
	authBinary := filepath.Join(temporaryDirectory, "auth")
	gatewayBinary := filepath.Join(temporaryDirectory, "gateway")
	buildBinary(t, repositoryRoot, ledgerBinary, "./ledger/cmd/ledger")
	buildBinary(t, repositoryRoot, authBinary, "./auth/cmd/auth")
	buildBinary(t, repositoryRoot, gatewayBinary, "./gateway/cmd/gateway")

	ledgerPort := freePort(t)
	authPort := freePort(t)
	gatewayPort := freePort(t)
	ledgerProcess := startProcess(
		t,
		repositoryRoot,
		ledgerBinary,
		"LEDGER_GRPC_PORT="+ledgerPort,
		"LEDGER_DATABASE_URL="+databaseURL,
	)
	t.Cleanup(func() { stopProcess(t, ledgerProcess) })
	waitForTCP(t, net.JoinHostPort("127.0.0.1", ledgerPort), ledgerProcess)

	authProcess := startProcess(
		t,
		repositoryRoot,
		authBinary,
		"AUTH_GRPC_PORT="+authPort,
		"AUTH_DATABASE_URL="+databaseURL,
		"AUTH_BCRYPT_COST=4",
		"AUTH_JWT_SECRET=0123456789abcdef0123456789abcdef",
	)
	t.Cleanup(func() { stopProcess(t, authProcess) })
	waitForTCP(t, net.JoinHostPort("127.0.0.1", authPort), authProcess)

	gatewayProcess := startProcess(
		t,
		repositoryRoot,
		gatewayBinary,
		"GATEWAY_HTTP_ADDR="+net.JoinHostPort("127.0.0.1", gatewayPort),
		"LEDGER_GRPC_HOST=127.0.0.1",
		"LEDGER_GRPC_PORT="+ledgerPort,
		"AUTH_GRPC_HOST=127.0.0.1",
		"AUTH_GRPC_PORT="+authPort,
		"AUTH_JWT_SECRET=0123456789abcdef0123456789abcdef",
	)
	t.Cleanup(func() { stopProcess(t, gatewayProcess) })

	baseURL := "http://" + net.JoinHostPort("127.0.0.1", gatewayPort)
	httpClient := &http.Client{Timeout: 5 * time.Second}
	waitForHTTP(t, httpClient, baseURL+"/ping", gatewayProcess)

	database, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}
	t.Cleanup(database.Close)
	cleanupData(t, database)
	t.Cleanup(func() { cleanupData(t, database) })

	assertStatus(t, httpClient, http.MethodGet, baseURL+"/ping", nil, http.StatusOK)
	assertStatus(
		t,
		httpClient,
		http.MethodPost,
		baseURL+"/auth/register",
		map[string]any{"email": testEmail, "password": "secure-password"},
		http.StatusCreated,
	)
	assertStatus(
		t,
		httpClient,
		http.MethodPost,
		baseURL+"/auth/register",
		map[string]any{"email": testEmail, "password": "secure-password"},
		http.StatusConflict,
	)
	authenticatedUserID, accessToken := assertLogin(t, httpClient, baseURL)
	assertStatus(
		t,
		httpClient,
		http.MethodPost,
		baseURL+"/auth/login",
		map[string]any{"email": testEmail, "password": "wrong-password"},
		http.StatusUnauthorized,
	)
	assertStatus(
		t,
		httpClient,
		http.MethodGet,
		baseURL+"/budgets",
		nil,
		http.StatusUnauthorized,
	)
	authenticatedClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: bearerTransport{
			token: accessToken,
			base:  http.DefaultTransport,
		},
	}
	assertStatus(
		t,
		authenticatedClient,
		http.MethodPut,
		baseURL+"/budgets/"+testCategory,
		map[string]any{"user_id": "another-user", "limit_amount": 3000},
		http.StatusBadRequest,
	)
	assertStatus(
		t,
		authenticatedClient,
		http.MethodPut,
		baseURL+"/budgets/"+testCategory,
		map[string]any{"limit_amount": 3000},
		http.StatusOK,
	)
	assertBudgetList(t, authenticatedClient, baseURL, authenticatedUserID)
	assertStatus(
		t,
		authenticatedClient,
		http.MethodPost,
		baseURL+"/transactions",
		map[string]any{
			"amount": 2000, "category": testCategory,
			"description": "end-to-end", "occurred_at": "2026-09-30T12:00:00Z",
		},
		http.StatusCreated,
	)
	assertTransactionList(t, authenticatedClient, baseURL, authenticatedUserID)
	assertSummary(t, authenticatedClient, baseURL, authenticatedUserID)
	assertCSVImport(t, authenticatedClient, baseURL, authenticatedUserID)
	assertCSVExport(t, authenticatedClient, baseURL)
	assertStatus(
		t,
		authenticatedClient,
		http.MethodPost,
		baseURL+"/transactions",
		map[string]any{
			"amount": 1500, "category": testCategory,
			"occurred_at": "2026-09-30T12:00:00Z",
		},
		http.StatusConflict,
	)
	assertRawStatus(
		t,
		authenticatedClient,
		http.MethodPost,
		baseURL+"/transactions",
		[]byte("{"),
		http.StatusBadRequest,
	)
}

func assertCSVImport(t *testing.T, client *http.Client, baseURL, userID string) {
	t.Helper()

	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		baseURL+"/transactions/import",
		strings.NewReader("amount,category,description,occurred_at\n500,"+testCategory+",imported,2026-09-30T12:30:00Z\n"),
	)
	if err != nil {
		t.Fatalf("create CSV import request: %v", err)
	}
	request.Header.Set("Content-Type", "text/csv")

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("import CSV: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected CSV import status %d, got %d", http.StatusOK, response.StatusCode)
	}

	var result struct {
		ImportedCount int `json:"imported_count"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("decode CSV import response: %v", err)
	}
	if result.ImportedCount != 1 {
		t.Fatalf("expected one imported transaction for %s, got %d", userID, result.ImportedCount)
	}
}

func assertCSVExport(t *testing.T, client *http.Client, baseURL string) {
	t.Helper()

	query := url.Values{
		"category": {testCategory},
		"from":     {"2026-09-01T00:00:00Z"},
		"to":       {"2026-10-01T00:00:00Z"},
	}
	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		baseURL+"/transactions/export?"+query.Encode(),
		http.NoBody,
	)
	if err != nil {
		t.Fatalf("create CSV export request: %v", err)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("export CSV: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected CSV export status %d, got %d", http.StatusOK, response.StatusCode)
	}
	if got := response.Header.Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Fatalf("unexpected CSV content type: %q", got)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read CSV export response: %v", err)
	}
	if !strings.Contains(string(body), "500,"+testCategory+",imported") {
		t.Fatalf("exported CSV does not contain imported transaction: %q", body)
	}
}

func assertLogin(t *testing.T, client *http.Client, baseURL string) (string, string) {
	t.Helper()

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(map[string]any{
		"email": testEmail, "password": "secure-password",
	}); err != nil {
		t.Fatalf("encode login request: %v", err)
	}
	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		baseURL+"/auth/login",
		&body,
	)
	if err != nil {
		t.Fatalf("create login request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected login status %d, got %d", http.StatusOK, response.StatusCode)
	}

	var result struct {
		UserID      string    `json:"user_id"`
		Email       string    `json:"email"`
		AccessToken string    `json:"access_token"`
		ExpiresAt   time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if result.UserID == "" || result.Email != testEmail ||
		result.AccessToken == "" || !result.ExpiresAt.After(time.Now()) {
		t.Fatalf("unexpected login response: %+v", result)
	}

	return result.UserID, result.AccessToken
}

func assertSummary(t *testing.T, client *http.Client, baseURL, userID string) {
	t.Helper()

	query := url.Values{
		"from": {"2026-09-01T00:00:00Z"},
		"to":   {"2026-10-01T00:00:00Z"},
	}
	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		baseURL+"/reports/summary?"+query.Encode(),
		http.NoBody,
	)
	if err != nil {
		t.Fatalf("create summary request: %v", err)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}

	var summary struct {
		UserID     string `json:"user_id"`
		TotalSpent int64  `json:"total_spent"`
		Categories []struct {
			Category         string `json:"category"`
			Spent            int64  `json:"spent"`
			BudgetLimit      int64  `json:"budget_limit"`
			BudgetConfigured bool   `json:"budget_configured"`
			Remaining        int64  `json:"remaining"`
			BudgetExceeded   bool   `json:"budget_exceeded"`
		} `json:"categories"`
	}
	if err := json.NewDecoder(response.Body).Decode(&summary); err != nil {
		t.Fatalf("decode summary response: %v", err)
	}
	if summary.UserID != userID || summary.TotalSpent != 2000 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if len(summary.Categories) != 1 {
		t.Fatalf("expected one summary category, got %d", len(summary.Categories))
	}
	category := summary.Categories[0]
	if category.Category != testCategory ||
		category.Spent != 2000 ||
		category.BudgetLimit != 3000 ||
		!category.BudgetConfigured ||
		category.Remaining != 1000 ||
		category.BudgetExceeded {
		t.Fatalf("unexpected summary category: %+v", category)
	}
}

func assertBudgetList(t *testing.T, client *http.Client, baseURL, userID string) {
	t.Helper()

	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		baseURL+"/budgets",
		http.NoBody,
	)
	if err != nil {
		t.Fatalf("create budget list request: %v", err)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("get budgets: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}

	var budgets []struct {
		UserID   string `json:"user_id"`
		Category string `json:"category"`
		Limit    int64  `json:"limit_amount"`
	}
	if err := json.NewDecoder(response.Body).Decode(&budgets); err != nil {
		t.Fatalf("decode budget list response: %v", err)
	}
	if len(budgets) != 1 {
		t.Fatalf("expected one budget, got %d", len(budgets))
	}
	budget := budgets[0]
	if budget.UserID != userID ||
		budget.Category != testCategory ||
		budget.Limit != 3000 {
		t.Fatalf("unexpected budget: %+v", budget)
	}
}

func assertTransactionList(t *testing.T, client *http.Client, baseURL, userID string) {
	t.Helper()

	query := url.Values{
		"category": {testCategory},
		"from":     {"2026-09-01T00:00:00Z"},
		"to":       {"2026-10-01T00:00:00Z"},
	}
	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		baseURL+"/transactions?"+query.Encode(),
		http.NoBody,
	)
	if err != nil {
		t.Fatalf("create transaction list request: %v", err)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("get transactions: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}

	var transactions []struct {
		UserID      string `json:"user_id"`
		Amount      int64  `json:"amount"`
		Category    string `json:"category"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(response.Body).Decode(&transactions); err != nil {
		t.Fatalf("decode transaction list response: %v", err)
	}
	if len(transactions) != 1 {
		t.Fatalf("expected one transaction, got %d", len(transactions))
	}
	transaction := transactions[0]
	if transaction.UserID != userID ||
		transaction.Amount != 2000 ||
		transaction.Category != testCategory ||
		transaction.Description != "end-to-end" {
		t.Fatalf("unexpected transaction: %+v", transaction)
	}
}

func buildBinary(t *testing.T, directory, output, packagePath string) {
	t.Helper()

	command := exec.Command("go", "build", "-o", output, packagePath)
	command.Dir = directory
	command.Stdout = os.Stderr
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		t.Fatalf("build %s: %v", packagePath, err)
	}
}

func startProcess(t *testing.T, directory, binary string, environment ...string) *runningProcess {
	t.Helper()

	command := exec.Command(binary)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
	command.Stdout = os.Stderr
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		t.Fatalf("start %s: %v", binary, err)
	}

	process := &runningProcess{command: command, done: make(chan error, 1)}
	go func() { process.done <- command.Wait() }()

	return process
}

func stopProcess(t *testing.T, process *runningProcess) {
	t.Helper()

	if err := process.command.Process.Signal(os.Interrupt); err != nil {
		if errorsIsProcessDone(err) {
			return
		}
		t.Errorf("interrupt process: %v", err)
	}

	select {
	case <-process.done:
	case <-time.After(5 * time.Second):
		if err := process.command.Process.Kill(); err != nil && !errorsIsProcessDone(err) {
			t.Errorf("kill process: %v", err)
		}
		<-process.done
	}
}

func errorsIsProcessDone(err error) bool {
	return errors.Is(err, os.ErrProcessDone) || errors.Is(err, syscall.ESRCH)
}

func freePort(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate port: %v", err)
	}
	port := fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port)
	if err := listener.Close(); err != nil {
		t.Fatalf("release port: %v", err)
	}

	return port
}

func waitForTCP(t *testing.T, address string, process *runningProcess) {
	t.Helper()
	waitUntilReady(t, process, func() bool {
		connection, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err != nil {
			return false
		}
		_ = connection.Close()

		return true
	})
}

func waitForHTTP(t *testing.T, client *http.Client, url string, process *runningProcess) {
	t.Helper()
	waitUntilReady(t, process, func() bool {
		response, err := client.Get(url)
		if err != nil {
			return false
		}
		_ = response.Body.Close()

		return response.StatusCode == http.StatusOK
	})
}

func waitUntilReady(t *testing.T, process *runningProcess, ready func() bool) {
	t.Helper()
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		if ready() {
			return
		}

		select {
		case err := <-process.done:
			t.Fatalf("process exited before readiness: %v", err)
		case <-timer.C:
			t.Fatal("timed out waiting for process readiness")
		case <-ticker.C:
		}
	}
}

func cleanupData(t *testing.T, database *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := database.Exec(
		ctx,
		"DELETE FROM transactions WHERE user_id IN (SELECT id FROM users WHERE email = $1)",
		testEmail,
	); err != nil {
		t.Fatalf("cleanup transactions: %v", err)
	}
	if _, err := database.Exec(
		ctx,
		"DELETE FROM budgets WHERE user_id IN (SELECT id FROM users WHERE email = $1) AND category = $2",
		testEmail,
		testCategory,
	); err != nil {
		t.Fatalf("cleanup budget: %v", err)
	}
	if _, err := database.Exec(ctx, "DELETE FROM users WHERE email = $1", testEmail); err != nil {
		t.Fatalf("cleanup user: %v", err)
	}
}

func assertStatus(
	t *testing.T,
	client *http.Client,
	method string,
	url string,
	body any,
	wantStatus int,
) {
	t.Helper()

	var encodedBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&encodedBody).Encode(body); err != nil {
			t.Fatalf("encode request: %v", err)
		}
	}
	assertRawStatus(t, client, method, url, encodedBody.Bytes(), wantStatus)
}

func assertRawStatus(
	t *testing.T,
	client *http.Client,
	method string,
	url string,
	body []byte,
	wantStatus int,
) {
	t.Helper()

	request, err := http.NewRequestWithContext(
		context.Background(),
		method,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != wantStatus {
		t.Fatalf("expected status %d, got %d", wantStatus, response.StatusCode)
	}
}

func requireEnv(t *testing.T, key string) string {
	t.Helper()

	value := os.Getenv(key)
	if value == "" {
		t.Fatalf("environment variable %s must be set", key)
	}

	return value
}
