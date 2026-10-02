package auth

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrAlreadyExists      = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrCanceled           = errors.New("request canceled")
	ErrDeadline           = errors.New("deadline exceeded")
	ErrUnavailable        = errors.New("service unavailable")
	ErrInternal           = errors.New("internal service error")
	ErrInvalidResponse    = errors.New("invalid Auth response")
)

func mapError(err error) error {
	grpcStatus, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("call Auth service: %w", ErrInternal)
	}

	var mappedError error
	switch grpcStatus.Code() {
	case codes.InvalidArgument:
		mappedError = ErrInvalidArgument
	case codes.AlreadyExists:
		mappedError = ErrAlreadyExists
	case codes.Unauthenticated:
		mappedError = ErrInvalidCredentials
	case codes.Canceled:
		mappedError = ErrCanceled
	case codes.DeadlineExceeded:
		mappedError = ErrDeadline
	case codes.Unavailable:
		mappedError = ErrUnavailable
	default:
		mappedError = ErrInternal
	}

	return fmt.Errorf("call Auth service: %w: %s", mappedError, grpcStatus.Message())
}
