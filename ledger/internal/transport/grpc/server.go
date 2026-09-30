package grpc

import (
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

var _ ledgerv1.LedgerServiceServer = (*Server)(nil)

type Server struct {
	ledgerv1.UnimplementedLedgerServiceServer

	service service.LedgerService
}

func New(ledgerService service.LedgerService) *Server {
	return &Server{service: ledgerService}
}
