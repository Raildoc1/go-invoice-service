package grpc

import (
	"fmt"
	apiservicepb "go-invoice-service/common/protocol/proto/apiservice"
	messageschedulerpb "go-invoice-service/common/protocol/proto/messagescheduler"
	validationpb "go-invoice-service/common/protocol/proto/validation"
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

type ValidationService interface {
	servers.ValidationService
}

type Config struct {
	Port uint16
}

type Server struct {
	cfg               Config
	invoiceService    InvoiceService
	outboxService     OutboxService
	validationService ValidationService
	server            *grpc.Server
}

func NewServer(
	cfg Config,
	invoiceService InvoiceService,
	outboxService OutboxService,
	validationService ValidationService,
) *Server {
	return &Server{
		invoiceService:    invoiceService,
		outboxService:     outboxService,
		validationService: validationService,
		server:            grpc.NewServer(),
		cfg:               cfg,
	}
}

func (s *Server) Run() error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", s.cfg.Port))
	if err != nil {
		return fmt.Errorf("failed to start listen: %w", err)
	}

	invoiceServer := servers.NewInvoiceServer(s.invoiceService)
	outboxServer := servers.NewOutboxServer(s.outboxService)
	validationServer := servers.NewValidationServer(s.validationService)

	apiservicepb.RegisterInvoiceStorageServer(s.server, invoiceServer)
	messageschedulerpb.RegisterOutboxStorageServer(s.server, outboxServer)
	validationpb.RegisterInvoiceStorageServer(s.server, validationServer)

	if err := s.server.Serve(listen); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) Shutdown() {
	s.server.GracefulStop()
}
