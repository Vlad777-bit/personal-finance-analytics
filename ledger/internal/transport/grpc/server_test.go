package grpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
	servicemocks "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service/mocks"
	grpctransport "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/transport/grpc"
	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func TestServer_CreateTransaction(t *testing.T) {
	t.Parallel()

	occurredAt := time.Date(2026, time.September, 30, 10, 0, 0, 0, time.UTC)
	createdAt := occurredAt.Add(time.Minute)
	ledgerService := servicemocks.NewLedgerService(t)
	ledgerService.EXPECT().CreateTransaction(
		mock.Anything,
		service.CreateTransactionInput{
			UserID:      "user-1",
			Amount:      1500,
			Category:    "food",
			Description: "lunch",
			OccurredAt:  occurredAt,
		},
	).Return(domain.Transaction{
		ID:          "transaction-1",
		UserID:      "user-1",
		Amount:      1500,
		Category:    "food",
		Description: "lunch",
		OccurredAt:  occurredAt,
		CreatedAt:   createdAt,
	}, nil)

	response, err := grpctransport.New(ledgerService).CreateTransaction(
		context.Background(),
		&ledgerv1.CreateTransactionRequest{
			UserId:      "user-1",
			Amount:      1500,
			Category:    "food",
			Description: "lunch",
			OccurredAt:  timestamppb.New(occurredAt),
		},
	)
	require.NoError(t, err)
	require.Equal(t, "transaction-1", response.GetTransaction().GetId())
	require.True(t, response.GetTransaction().GetOccurredAt().AsTime().Equal(occurredAt))
	require.True(t, response.GetTransaction().GetCreatedAt().AsTime().Equal(createdAt))
}

func TestServer_CreateBudget(t *testing.T) {
	t.Parallel()

	ledgerService := servicemocks.NewLedgerService(t)
	ledgerService.EXPECT().CreateBudget(
		mock.Anything,
		service.CreateBudgetInput{UserID: "user-1", Category: "food", Limit: 50000},
	).Return(domain.Budget{
		ID: "budget-1", UserID: "user-1", Category: "food", Limit: 50000,
	}, nil)

	response, err := grpctransport.New(ledgerService).CreateBudget(
		context.Background(),
		&ledgerv1.CreateBudgetRequest{
			UserId: "user-1", Category: "food", LimitAmount: 50000,
		},
	)
	require.NoError(t, err)
	require.Equal(t, "budget-1", response.GetBudget().GetId())
	require.Equal(t, int64(50000), response.GetBudget().GetLimitAmount())
}

func TestServer_GetTransactions(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	createdAt := from.Add(time.Hour)

	ledgerService := servicemocks.NewLedgerService(t)
	ledgerService.EXPECT().GetTransactions(
		mock.Anything,
		service.GetTransactionsInput{
			UserID:   "user-1",
			Category: "food",
			From:     from,
			To:       to,
		},
	).Return([]domain.Transaction{
		{
			ID:          "transaction-1",
			UserID:      "user-1",
			Amount:      1500,
			Category:    "food",
			Description: "lunch",
			OccurredAt:  from,
			CreatedAt:   createdAt,
		},
	}, nil)

	response, err := grpctransport.New(ledgerService).GetTransactions(
		t.Context(),
		&ledgerv1.GetTransactionsRequest{
			UserId:   "user-1",
			Category: "food",
			From:     timestamppb.New(from),
			To:       timestamppb.New(to),
		},
	)
	require.NoError(t, err)
	require.Len(t, response.GetTransactions(), 1)
	require.Equal(t, "transaction-1", response.GetTransactions()[0].GetId())
	require.True(
		t,
		response.GetTransactions()[0].GetCreatedAt().AsTime().Equal(createdAt),
	)
}

func TestServer_GetBudgets(t *testing.T) {
	t.Parallel()

	ledgerService := servicemocks.NewLedgerService(t)
	ledgerService.EXPECT().GetBudgets(
		mock.Anything,
		service.GetBudgetsInput{UserID: "user-1"},
	).Return([]domain.Budget{
		{ID: "budget-1", UserID: "user-1", Category: "food", Limit: 50000},
	}, nil)

	response, err := grpctransport.New(ledgerService).GetBudgets(
		t.Context(),
		&ledgerv1.GetBudgetsRequest{UserId: "user-1"},
	)
	require.NoError(t, err)
	require.Len(t, response.GetBudgets(), 1)
	require.Equal(t, "budget-1", response.GetBudgets()[0].GetId())
	require.Equal(t, int64(50000), response.GetBudgets()[0].GetLimitAmount())
}

func TestServer_RejectsNilRequest(t *testing.T) {
	t.Parallel()

	server := grpctransport.New(servicemocks.NewLedgerService(t))

	_, transactionErr := server.CreateTransaction(context.Background(), nil)
	_, budgetErr := server.CreateBudget(context.Background(), nil)
	_, getTransactionsErr := server.GetTransactions(context.Background(), nil)
	_, getBudgetsErr := server.GetBudgets(context.Background(), nil)

	require.Equal(t, codes.InvalidArgument, status.Code(transactionErr))
	require.Equal(t, codes.InvalidArgument, status.Code(budgetErr))
	require.Equal(t, codes.InvalidArgument, status.Code(getTransactionsErr))
	require.Equal(t, codes.InvalidArgument, status.Code(getBudgetsErr))
}

func TestServer_GetTransactionsRejectsInvalidTimestamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		request *ledgerv1.GetTransactionsRequest
	}{
		{
			name: "invalid from",
			request: &ledgerv1.GetTransactionsRequest{
				From: &timestamppb.Timestamp{Seconds: 253402300800},
			},
		},
		{
			name: "invalid to",
			request: &ledgerv1.GetTransactionsRequest{
				To: &timestamppb.Timestamp{Seconds: 253402300800},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := grpctransport.New(servicemocks.NewLedgerService(t))
			_, err := server.GetTransactions(t.Context(), tt.request)

			require.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestServer_RejectsInvalidTimestamp(t *testing.T) {
	t.Parallel()

	server := grpctransport.New(servicemocks.NewLedgerService(t))

	_, err := server.CreateTransaction(
		context.Background(),
		&ledgerv1.CreateTransactionRequest{
			OccurredAt: &timestamppb.Timestamp{Seconds: 253402300800},
		},
	)

	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
