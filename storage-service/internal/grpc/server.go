package grpc

import (
	"fmt"
	pb "go-invoice-service/common/protocol/proto/apiservice/storage"
	"google.golang.org/grpc"
	"net"
)

var _ pb.InvoiceStorageServer = (*Server)(nil)

type Controller interface {
}

type Config struct {
	Port uint16
}

type Server struct {
	pb.UnimplementedInvoiceStorageServer
	cfg        Config
	controller Controller
	server     *grpc.Server
}

func NewServer(cfg Config, controller Controller) *Server {
	return &Server{
		controller: controller,
		server:     grpc.NewServer(),
		cfg:        cfg,
	}
}

func (s *Server) Run() error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", s.cfg.Port))
	if err != nil {
		return fmt.Errorf("failed to start listen: %w", err)
	}

	invoiceService := grpcservers.NewUpdateMetricsServer(s.controller)

	pb.RegisterInvoiceStorageServer(s.server, invoiceService)

	if err := s.server.Serve(listen); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) Shutdown() {
	s.server.GracefulStop()
}
