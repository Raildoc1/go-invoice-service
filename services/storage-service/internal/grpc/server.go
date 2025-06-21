package grpc

import (
	"context"
	"fmt"
	pb "go-invoice-service/common/protocol/proto/apiservice"
	"google.golang.org/grpc"
	"net"
	"storage-service/internal/dto"
	"storage-service/internal/grpc/servers"
)

type InvoiceService interface {
	AddNew(ctx context.Context, invoice dto.Invoice) error
}

type OutboxService interface{}

type Config struct {
	Port uint16
}

type Server struct {
	cfg            Config
	invoiceService InvoiceService
	server         *grpc.Server
}

func NewServer(cfg Config, invoiceService InvoiceService) *Server {
	return &Server{
		invoiceService: invoiceService,
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

	pb.RegisterInvoiceStorageServer(s.server, invoiceService)

	if err := s.server.Serve(listen); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) Shutdown() {
	s.server.GracefulStop()
}
