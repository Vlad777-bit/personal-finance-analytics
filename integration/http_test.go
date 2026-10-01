//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	testUserID   = "66666666-6666-6666-6666-666666666666"
	testCategory = "http-e2e"
)

type runningProcess struct {
	command *exec.Cmd
	done    chan error
}

func TestGatewayLedgerHTTP(t *testing.T) {
	databaseURL := requireEnv(t, "LEDGER_DATABASE_URL")
	repositoryRoot := filepath.Clean("..")
	temporaryDirectory := t.TempDir()

	ledgerBinary := filepath.Join(temporaryDirectory, "ledger")
	gatewayBinary := filepath.Join(temporaryDirectory, "gateway")
	buildBinary(t, repositoryRoot, ledgerBinary, "./ledger/cmd/ledger")
	buildBinary(t, repositoryRoot, gatewayBinary, "./gateway/cmd/gateway")

	ledgerPort := freePort(t)
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

	gatewayProcess := startProcess(
		t,
		repositoryRoot,
		gatewayBinary,
		"GATEWAY_HTTP_ADDR="+net.JoinHostPort("127.0.0.1", gatewayPort),
		"LEDGER_GRPC_HOST=127.0.0.1",
		"LEDGER_GRPC_PORT="+ledgerPort,
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
		http.MethodPut,
		baseURL+"/budgets/"+testCategory,
		map[string]any{"user_id": testUserID, "limit_amount": 3000},
		http.StatusOK,
	)
	assertBudgetList(t, httpClient, baseURL)
	assertStatus(
		t,
		httpClient,
		http.MethodPost,
		baseURL+"/transactions",
		map[string]any{
			"user_id": testUserID, "amount": 2000, "category": testCategory,
			"description": "end-to-end", "occurred_at": "2026-09-30T12:00:00Z",
		},
		http.StatusCreated,
	)
	assertTransactionList(t, httpClient, baseURL)
	assertStatus(
		t,
		httpClient,
		http.MethodPost,
		baseURL+"/transactions",
		map[string]any{
			"user_id": testUserID, "amount": 1500, "category": testCategory,
			"occurred_at": "2026-09-30T12:00:00Z",
		},
		http.StatusConflict,
	)
	assertRawStatus(
		t,
		httpClient,
		http.MethodPost,
		baseURL+"/transactions",
		[]byte("{"),
		http.StatusBadRequest,
	)
}

func assertBudgetList(t *testing.T, client *http.Client, baseURL string) {
	t.Helper()

	query := url.Values{"user_id": {testUserID}}
	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		baseURL+"/budgets?"+query.Encode(),
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
	if budget.UserID != testUserID ||
		budget.Category != testCategory ||
		budget.Limit != 3000 {
		t.Fatalf("unexpected budget: %+v", budget)
	}
}

func assertTransactionList(t *testing.T, client *http.Client, baseURL string) {
	t.Helper()

	query := url.Values{
		"user_id":  {testUserID},
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
	if transaction.UserID != testUserID ||
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
	if _, err := database.Exec(ctx, "DELETE FROM transactions WHERE user_id = $1", testUserID); err != nil {
		t.Fatalf("cleanup transactions: %v", err)
	}
	if _, err := database.Exec(
		ctx,
		"DELETE FROM budgets WHERE user_id = $1 AND category = $2",
		testUserID,
		testCategory,
	); err != nil {
		t.Fatalf("cleanup budget: %v", err)
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
