package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/service"
	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

func (s *Server) Refresh(ctx context.Context, request *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	if request == nil || request.GetRefreshToken() == "" {
		return nil, invalidRequestError()
	}
	result, err := s.service.Refresh(ctx, service.RefreshInput{RefreshToken: request.GetRefreshToken()})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &authv1.RefreshResponse{AccessToken: result.Value, ExpiresAt: timestamppb.New(result.ExpiresAt)}, nil
}
