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
	getTransactions func(
		context.Context,
		*ledgerv1.GetTransactionsRequest,
	) (*ledgerv1.GetTransactionsResponse, error)
	getBudgets func(
		context.Context,
		*ledgerv1.GetBudgetsRequest,
	) (*ledgerv1.GetBudgetsResponse, error)
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
	ctx context.Context,
	request *ledgerv1.GetTransactionsRequest,
	_ ...grpc.CallOption,
) (*ledgerv1.GetTransactionsResponse, error) {
	return f.getTransactions(ctx, request)
}

func (f *fakeLedgerServiceClient) GetBudgets(
	ctx context.Context,
	request *ledgerv1.GetBudgetsRequest,
	_ ...grpc.CallOption,
) (*ledgerv1.GetBudgetsResponse, error) {
	return f.getBudgets(ctx, request)
}

func (f *fakeLedgerServiceClient) GetSummary(
	_ context.Context,
	_ *ledgerv1.GetSummaryRequest,
	_ ...grpc.CallOption,
) (*ledgerv1.GetSummaryResponse, error) {
	panic("unexpected GetSummary call")
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

func TestClient_GetTransactions(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	createdAt := from.Add(time.Hour)

	tests := []struct {
		name    string
		call    func(*testing.T, *ledgerv1.GetTransactionsRequest) (*ledgerv1.GetTransactionsResponse, error)
		want    []Transaction
		wantErr error
	}{
		{
			name: "success",
			call: func(t *testing.T, request *ledgerv1.GetTransactionsRequest) (*ledgerv1.GetTransactionsResponse, error) {
				t.Helper()
				require.Equal(t, "user-1", request.GetUserId())
				require.Equal(t, "food", request.GetCategory())
				require.True(t, request.GetFrom().AsTime().Equal(from))
				require.True(t, request.GetTo().AsTime().Equal(to))

				return &ledgerv1.GetTransactionsResponse{
					Transactions: []*ledgerv1.Transaction{
						{
							Id: "transaction-1", UserId: "user-1", Amount: 1500,
							Category: "food", Description: "lunch",
							OccurredAt: timestamppb.New(from),
							CreatedAt:  timestamppb.New(createdAt),
						},
					},
				}, nil
			},
			want: []Transaction{
				{
					ID: "transaction-1", UserID: "user-1", Amount: 1500,
					Category: "food", Description: "lunch",
					OccurredAt: from, CreatedAt: createdAt,
				},
			},
		},
		{
			name: "empty result",
			call: func(_ *testing.T, _ *ledgerv1.GetTransactionsRequest) (*ledgerv1.GetTransactionsResponse, error) {
				return &ledgerv1.GetTransactionsResponse{}, nil
			},
			want: []Transaction{},
		},
		{
			name: "grpc error",
			call: func(_ *testing.T, _ *ledgerv1.GetTransactionsRequest) (*ledgerv1.GetTransactionsResponse, error) {
				return nil, status.Error(codes.InvalidArgument, "invalid period")
			},
			wantErr: ErrInvalidArgument,
		},
		{
			name: "nil response",
			call: func(_ *testing.T, _ *ledgerv1.GetTransactionsRequest) (*ledgerv1.GetTransactionsResponse, error) {
				return nil, nil
			},
			wantErr: ErrInvalidResponse,
		},
		{
			name: "invalid transaction",
			call: func(_ *testing.T, _ *ledgerv1.GetTransactionsRequest) (*ledgerv1.GetTransactionsResponse, error) {
				return &ledgerv1.GetTransactionsResponse{
					Transactions: []*ledgerv1.Transaction{nil},
				}, nil
			},
			wantErr: ErrInvalidResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &Client{service: &fakeLedgerServiceClient{
				getTransactions: func(
					_ context.Context,
					request *ledgerv1.GetTransactionsRequest,
				) (*ledgerv1.GetTransactionsResponse, error) {
					return tt.call(t, request)
				},
			}}

			got, err := client.GetTransactions(t.Context(), GetTransactionsInput{
				UserID: "user-1", Category: "food", From: from, To: to,
			})

			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestClient_GetBudgets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		call    func(*testing.T, *ledgerv1.GetBudgetsRequest) (*ledgerv1.GetBudgetsResponse, error)
		want    []Budget
		wantErr error
	}{
		{
			name: "success",
			call: func(t *testing.T, request *ledgerv1.GetBudgetsRequest) (*ledgerv1.GetBudgetsResponse, error) {
				t.Helper()
				require.Equal(t, "user-1", request.GetUserId())

				return &ledgerv1.GetBudgetsResponse{
					Budgets: []*ledgerv1.Budget{
						{
							Id: "budget-1", UserId: "user-1",
							Category: "food", LimitAmount: 50000,
						},
					},
				}, nil
			},
			want: []Budget{
				{ID: "budget-1", UserID: "user-1", Category: "food", Limit: 50000},
			},
		},
		{
			name: "empty result",
			call: func(_ *testing.T, _ *ledgerv1.GetBudgetsRequest) (*ledgerv1.GetBudgetsResponse, error) {
				return &ledgerv1.GetBudgetsResponse{}, nil
			},
			want: []Budget{},
		},
		{
			name: "grpc error",
			call: func(_ *testing.T, _ *ledgerv1.GetBudgetsRequest) (*ledgerv1.GetBudgetsResponse, error) {
				return nil, status.Error(codes.InvalidArgument, "user id is required")
			},
			wantErr: ErrInvalidArgument,
		},
		{
			name: "nil response",
			call: func(_ *testing.T, _ *ledgerv1.GetBudgetsRequest) (*ledgerv1.GetBudgetsResponse, error) {
				return nil, nil
			},
			wantErr: ErrInvalidResponse,
		},
		{
			name: "invalid budget",
			call: func(_ *testing.T, _ *ledgerv1.GetBudgetsRequest) (*ledgerv1.GetBudgetsResponse, error) {
				return &ledgerv1.GetBudgetsResponse{
					Budgets: []*ledgerv1.Budget{nil},
				}, nil
			},
			wantErr: ErrInvalidResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &Client{service: &fakeLedgerServiceClient{
				getBudgets: func(
					_ context.Context,
					request *ledgerv1.GetBudgetsRequest,
				) (*ledgerv1.GetBudgetsResponse, error) {
					return tt.call(t, request)
				},
			}}

			got, err := client.GetBudgets(
				t.Context(),
				GetBudgetsInput{UserID: "user-1"},
			)

			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
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
