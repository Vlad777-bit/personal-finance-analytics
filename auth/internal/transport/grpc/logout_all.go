package grpc

import (
	"context"

	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

func (s *Server) LogoutAll(ctx context.Context, request *authv1.LogoutAllRequest) (*authv1.LogoutAllResponse, error) {
	if request == nil || request.GetUserId() == "" {
		return nil, invalidRequestError()
	}
	if err := s.service.LogoutAll(ctx, request.GetUserId()); err != nil {
		return nil, mapServiceError(err)
	}
	return &authv1.LogoutAllResponse{}, nil
}
