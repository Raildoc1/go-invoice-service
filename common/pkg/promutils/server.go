package promutils

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

type PrometheusConfig struct {
	PortToListen    uint16
	ShutdownTimeout time.Duration
}

type Server struct {
	cfg PrometheusConfig
	srv *http.Server
}

func NewServer(
	cfg PrometheusConfig) *Server {
	return &Server{
		cfg: cfg,
		srv: nil,
	}
}

func (s *Server) Run() error {
	s.srv = &http.Server{
		Addr:    fmt.Sprintf(":%v", s.cfg.PortToListen),
		Handler: s.createMux(),
	}
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

func (s *Server) createMux() *chi.Mux {
	router := chi.NewRouter()

	router.Handle("/metrics", promhttp.Handler())

	return router
}
