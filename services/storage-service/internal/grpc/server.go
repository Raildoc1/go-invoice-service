package grpc

import (
	"fmt"
	apiservicepb "go-invoice-service/common/protocol/proto/apiservice"
	messageschedulerpb "go-invoice-service/common/protocol/proto/messagescheduler"
	"google.golang.org/grpc"
	"net"
	"storage-service/internal/grpc/servers"
)

type InvoiceService interface {
	servers.InvoiceService
}

type OutboxService interface {
	servers.OutboxService
}

type Config struct {
	Port uint16
}

type Server struct {
	cfg            Config
	invoiceService InvoiceService
	outboxService  OutboxService
	server         *grpc.Server
}

func NewServer(cfg Config, invoiceService InvoiceService, outboxService OutboxService) *Server {
	return &Server{
		invoiceService: invoiceService,
		outboxService:  outboxService,
		server:         grpc.NewServer(),
		cfg:            cfg,
	}
}

func (s *Server) Run() error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", s.cfg.Port))
	if err != nil {
		return fmt.Errorf("failed to start listen: %w", err)
	}

	invoiceService := servers.NewInvoiceServer(s.invoiceService)
	outboxService := servers.NewOutboxServer(s.outboxService)

	apiservicepb.RegisterInvoiceStorageServer(s.server, invoiceService)
	messageschedulerpb.RegisterOutboxStorageServer(s.server, outboxService)

	if err := s.server.Serve(listen); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) Shutdown() {
	s.server.GracefulStop()
}
