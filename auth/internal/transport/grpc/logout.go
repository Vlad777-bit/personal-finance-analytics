package grpc

import (
	"context"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/service"
	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

func (s *Server) Logout(ctx context.Context, request *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if request == nil || request.GetRefreshToken() == "" {
		return nil, invalidRequestError()
	}
	if err := s.service.Logout(ctx, service.RefreshInput{RefreshToken: request.GetRefreshToken()}); err != nil {
		return nil, mapServiceError(err)
	}
	return &authv1.LogoutResponse{}, nil
}
