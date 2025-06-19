package http

import (
	"compress/gzip"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"go-invoice-service/api-service/internal/http/handlers"
	"go-invoice-service/common/pkg/http/middleware"
	"go-invoice-service/common/pkg/logging"
	"net/http"
)

type Server struct {
	srv          *http.Server
	cfg          Config
	jwtTokenAuth *jwtauth.JWTAuth
	logger       *logging.ZapLogger
}

func NewServer(cfg Config, tokenAuth *jwtauth.JWTAuth, logger *logging.ZapLogger) *Server {
	return &Server{
		srv:          nil,
		cfg:          cfg,
		jwtTokenAuth: tokenAuth,
		logger:       logger,
	}
}

func (s *Server) Run() error {
	s.srv = &http.Server{
		Addr:    s.cfg.ServerAddress,
		Handler: s.createMux(),
	}
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) createMux() *chi.Mux {
	router := chi.NewRouter()

	// middleware
	panicRecover := middleware.NewPanicRecover(s.logger)
	loggerContextMiddleware := middleware.NewLogging()
	requestDecompression := middleware.NewRequestDecompressor(s.logger)
	responseCompression := middleware.NewResponseCompressor(s.logger, gzip.BestSpeed)

	//handlers
	invoiceHandler := handlers.NewInvoice()
	authHandler := handlers.NewAuth()

	// router
	router.Use(loggerContextMiddleware.CreateHandler)
	router.Use(panicRecover.CreateHandler)
	router.Route("/api/user/", func(router chi.Router) {
		router.Post("/register", authHandler.Register)
		router.Post("/login", authHandler.Login)
		router.With(
			jwtauth.Verifier(s.jwtTokenAuth),
			jwtauth.Authenticator(s.jwtTokenAuth),
			requestDecompression.CreateHandler,
			responseCompression.CreateHandler,
		).Route("/invoice/", func(router chi.Router) {
			router.Post("/create", invoiceHandler.UploadNew)
			router.Get("/get", invoiceHandler.Get)
		})
	})

	return router
}
