package ledger

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

type fakeLedgerServiceClient struct {
	createTransaction func(
		context.Context,
		*ledgerv1.CreateTransactionRequest,
	) (*ledgerv1.CreateTransactionResponse, error)
	createBudget func(
		context.Context,
		*ledgerv1.CreateBudgetRequest,
	) (*ledgerv1.CreateBudgetResponse, error)
}

func (f *fakeLedgerServiceClient) CreateTransaction(
	ctx context.Context,
	request *ledgerv1.CreateTransactionRequest,
	_ ...grpc.CallOption,
) (*ledgerv1.CreateTransactionResponse, error) {
	return f.createTransaction(ctx, request)
}

func (f *fakeLedgerServiceClient) CreateBudget(
	ctx context.Context,
	request *ledgerv1.CreateBudgetRequest,
	_ ...grpc.CallOption,
) (*ledgerv1.CreateBudgetResponse, error) {
	return f.createBudget(ctx, request)
}

func (f *fakeLedgerServiceClient) GetTransactions(
	context.Context,
	*ledgerv1.GetTransactionsRequest,
	...grpc.CallOption,
) (*ledgerv1.GetTransactionsResponse, error) {
	panic("unexpected GetTransactions call")
}

func TestClient_CreateTransaction(t *testing.T) {
	t.Parallel()

	occurredAt := time.Date(2026, time.September, 30, 12, 0, 0, 0, time.UTC)
	createdAt := occurredAt.Add(time.Second)
	client := &Client{service: &fakeLedgerServiceClient{
		createTransaction: func(
			_ context.Context,
			request *ledgerv1.CreateTransactionRequest,
		) (*ledgerv1.CreateTransactionResponse, error) {
			require.Equal(t, "user-1", request.GetUserId())
			require.True(t, request.GetOccurredAt().AsTime().Equal(occurredAt))

			return &ledgerv1.CreateTransactionResponse{
				Transaction: &ledgerv1.Transaction{
					Id: "transaction-1", UserId: "user-1", Amount: 1500,
					Category: "food", Description: "lunch",
					OccurredAt: timestamppb.New(occurredAt),
					CreatedAt:  timestamppb.New(createdAt),
				},
			}, nil
		},
	}}

	transaction, err := client.CreateTransaction(
		context.Background(),
		CreateTransactionInput{
			UserID: "user-1", Amount: 1500, Category: "food",
			Description: "lunch", OccurredAt: occurredAt,
		},
	)
	require.NoError(t, err)
	require.Equal(t, "transaction-1", transaction.ID)
	require.True(t, transaction.CreatedAt.Equal(createdAt))
}

func TestClient_CreateBudget(t *testing.T) {
	t.Parallel()

	client := &Client{service: &fakeLedgerServiceClient{
		createBudget: func(
			_ context.Context,
			request *ledgerv1.CreateBudgetRequest,
		) (*ledgerv1.CreateBudgetResponse, error) {
			require.Equal(t, int64(50000), request.GetLimitAmount())

			return &ledgerv1.CreateBudgetResponse{
				Budget: &ledgerv1.Budget{
					Id: "budget-1", UserId: "user-1", Category: "food", LimitAmount: 50000,
				},
			}, nil
		},
	}}

	budget, err := client.CreateBudget(
		context.Background(),
		CreateBudgetInput{UserID: "user-1", Category: "food", Limit: 50000},
	)
	require.NoError(t, err)
	require.Equal(t, "budget-1", budget.ID)
	require.Equal(t, int64(50000), budget.Limit)
}

func TestMapError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code codes.Code
		want error
	}{
		{name: "invalid argument", code: codes.InvalidArgument, want: ErrInvalidArgument},
		{name: "budget exceeded", code: codes.FailedPrecondition, want: ErrBudgetExceeded},
		{name: "not found", code: codes.NotFound, want: ErrNotFound},
		{name: "canceled", code: codes.Canceled, want: ErrCanceled},
		{name: "deadline", code: codes.DeadlineExceeded, want: ErrDeadline},
		{name: "unavailable", code: codes.Unavailable, want: ErrUnavailable},
		{name: "internal", code: codes.Internal, want: ErrInternal},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := mapError(status.Error(test.code, "details"))
			require.ErrorIs(t, err, test.want)
		})
	}

	require.ErrorIs(t, mapError(errors.New("transport error")), ErrInternal)
}
