package ledger

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrBudgetExceeded  = errors.New("budget exceeded")
	ErrNotFound        = errors.New("not found")
	ErrCanceled        = errors.New("request canceled")
	ErrDeadline        = errors.New("deadline exceeded")
	ErrUnavailable     = errors.New("service unavailable")
	ErrInternal        = errors.New("internal service error")
	ErrInvalidResponse = errors.New("invalid Ledger response")
)

func mapError(err error) error {
	grpcStatus, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("call Ledger service: %w", ErrInternal)
	}

	var mappedError error
	switch grpcStatus.Code() {
	case codes.InvalidArgument:
		mappedError = ErrInvalidArgument
	case codes.FailedPrecondition:
		mappedError = ErrBudgetExceeded
	case codes.NotFound:
		mappedError = ErrNotFound
	case codes.Canceled:
		mappedError = ErrCanceled
	case codes.DeadlineExceeded:
		mappedError = ErrDeadline
	case codes.Unavailable:
		mappedError = ErrUnavailable
	default:
		mappedError = ErrInternal
	}

	return fmt.Errorf("call Ledger service: %w: %s", mappedError, grpcStatus.Message())
}
