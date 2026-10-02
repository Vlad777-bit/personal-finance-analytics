package grpc

import (
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/service"
	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

var _ authv1.AuthServiceServer = (*Server)(nil)

type Server struct {
	authv1.UnimplementedAuthServiceServer

	service service.AuthService
}

func New(authService service.AuthService) *Server {
	return &Server{service: authService}
}
