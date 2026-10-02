package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/service"
	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

func (s *Server) Register(
	ctx context.Context,
	request *authv1.RegisterRequest,
) (*authv1.RegisterResponse, error) {
	if request == nil {
		return nil, invalidRequestError()
	}

	user, err := s.service.Register(ctx, service.RegisterInput{
		Email:    request.GetEmail(),
		Password: request.GetPassword(),
	})
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &authv1.RegisterResponse{
		User: &authv1.User{
			Id:        user.ID,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
		},
	}, nil
}
