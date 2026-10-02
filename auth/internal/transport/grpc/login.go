package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/service"
	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

func (s *Server) Login(
	ctx context.Context,
	request *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {
	if request == nil {
		return nil, invalidRequestError()
	}

	result, err := s.service.Login(ctx, service.LoginInput{
		Email:    request.GetEmail(),
		Password: request.GetPassword(),
	})
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &authv1.LoginResponse{
		UserId:      result.UserID,
		Email:       result.Email,
		AccessToken: result.AccessToken.Value,
		ExpiresAt:   timestamppb.New(result.AccessToken.ExpiresAt),
	}, nil
}
