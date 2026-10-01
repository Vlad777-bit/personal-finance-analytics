//go:build integration

package app

import (
	"context"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/testhelper"
	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

const (
	testGRPCUserID   = "33333333-3333-3333-3333-333333333333"
	testGRPCCategory = "grpc-integration"
)

func TestLedgerGRPC(t *testing.T) {
	databaseURL := testhelper.RequireEnv(t, "LEDGER_DATABASE_URL")
	redisAddress := integrationRedisAddress()
	applicationContext, cancelApplication := context.WithCancel(context.Background())

	application, err := New(
		applicationContext,
		databaseURL,
		redisAddress,
		5*time.Minute,
		"127.0.0.1:0",
		slog.Default(),
	)
	require.NoError(t, err)

	runErrors := make(chan error, 1)
	go func() {
		runErrors <- application.Run(applicationContext)
	}()

	t.Cleanup(func() {
		cancelApplication()
		require.NoError(t, <-runErrors)

		shutdownContext, cancelShutdown := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancelShutdown()

		require.NoError(t, application.Shutdown(shutdownContext))
	})

	connection, err := grpc.NewClient(
		application.listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, connection.Close())
	})

	testhelper.CleanupTransactions(
		t,
		context.Background(),
		application.database,
		testGRPCUserID,
	)
	testhelper.CleanupBudget(
		t,
		context.Background(),
		application.database,
		testGRPCUserID,
		testGRPCCategory,
	)
	t.Cleanup(func() {
		testhelper.CleanupTransactions(
			t,
			context.Background(),
			application.database,
			testGRPCUserID,
		)
		testhelper.CleanupBudget(
			t,
			context.Background(),
			application.database,
			testGRPCUserID,
			testGRPCCategory,
		)
	})

	client := ledgerv1.NewLedgerServiceClient(connection)
	rpcContext, cancelRPC := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelRPC()

	t.Run("create budget", func(t *testing.T) {
		response, callErr := client.CreateBudget(
			rpcContext,
			&ledgerv1.CreateBudgetRequest{
				UserId:      testGRPCUserID,
				Category:    testGRPCCategory,
				LimitAmount: 3000,
			},
		)
		require.NoError(t, callErr)
		require.NotEmpty(t, response.GetBudget().GetId())
		require.Equal(t, int64(3000), response.GetBudget().GetLimitAmount())
	})

	t.Run("get budgets", func(t *testing.T) {
		response, callErr := client.GetBudgets(
			rpcContext,
			&ledgerv1.GetBudgetsRequest{UserId: testGRPCUserID},
		)
		require.NoError(t, callErr)
		require.Len(t, response.GetBudgets(), 1)
		require.Equal(t, testGRPCCategory, response.GetBudgets()[0].GetCategory())
		require.Equal(t, int64(3000), response.GetBudgets()[0].GetLimitAmount())
	})

	occurredAt := time.Date(2026, time.September, 30, 12, 0, 0, 0, time.UTC)

	t.Run("create transaction", func(t *testing.T) {
		response, callErr := client.CreateTransaction(
			rpcContext,
			&ledgerv1.CreateTransactionRequest{
				UserId:      testGRPCUserID,
				Amount:      2000,
				Category:    testGRPCCategory,
				Description: "integration transaction",
				OccurredAt:  timestamppb.New(occurredAt),
			},
		)
		require.NoError(t, callErr)
		require.NotEmpty(t, response.GetTransaction().GetId())
		require.Equal(t, int64(2000), response.GetTransaction().GetAmount())
	})

	t.Run("get transactions", func(t *testing.T) {
		response, callErr := client.GetTransactions(
			rpcContext,
			&ledgerv1.GetTransactionsRequest{
				UserId:   testGRPCUserID,
				Category: testGRPCCategory,
				From:     timestamppb.New(occurredAt.Add(-time.Hour)),
				To:       timestamppb.New(occurredAt.Add(time.Hour)),
			},
		)
		require.NoError(t, callErr)
		require.Len(t, response.GetTransactions(), 1)
		require.Equal(
			t,
			"integration transaction",
			response.GetTransactions()[0].GetDescription(),
		)
	})

	t.Run("get summary", func(t *testing.T) {
		response, callErr := client.GetSummary(
			rpcContext,
			&ledgerv1.GetSummaryRequest{
				UserId: testGRPCUserID,
				From:   timestamppb.New(occurredAt.Add(-time.Hour)),
				To:     timestamppb.New(occurredAt.Add(time.Hour)),
			},
		)
		require.NoError(t, callErr)
		require.Equal(t, int64(2000), response.GetSummary().GetTotalSpent())
		require.Len(t, response.GetSummary().GetCategories(), 1)
		category := response.GetSummary().GetCategories()[0]
		require.Equal(t, testGRPCCategory, category.GetCategory())
		require.Equal(t, int64(1000), category.GetRemaining())
		require.True(t, category.GetBudgetConfigured())
		require.False(t, category.GetBudgetExceeded())
	})

	t.Run("budget exceeded", func(t *testing.T) {
		_, callErr := client.CreateTransaction(
			rpcContext,
			&ledgerv1.CreateTransactionRequest{
				UserId:     testGRPCUserID,
				Amount:     1500,
				Category:   testGRPCCategory,
				OccurredAt: timestamppb.New(occurredAt),
			},
		)
		require.Equal(t, codes.FailedPrecondition, status.Code(callErr))
	})

	t.Run("invalid request", func(t *testing.T) {
		_, callErr := client.CreateBudget(
			rpcContext,
			&ledgerv1.CreateBudgetRequest{},
		)
		require.Equal(t, codes.InvalidArgument, status.Code(callErr))
	})
}

func integrationRedisAddress() string {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	return net.JoinHostPort(host, port)
}
