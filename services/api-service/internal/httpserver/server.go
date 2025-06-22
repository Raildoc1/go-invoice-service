package httpserver

import (
	"compress/gzip"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"go-invoice-service/api-service/internal/httpserver/handlers"
	"go-invoice-service/api-service/internal/httpserver/middleware"
	commonMiddleware "go-invoice-service/common/pkg/http/middleware"
	"go-invoice-service/common/pkg/logging"
	"net/http"
)

type StorageService interface {
	handlers.StorageService
}

type Server struct {
	srv            *http.Server
	cfg            Config
	jwtTokenAuth   *jwtauth.JWTAuth
	storageService StorageService
	logger         *logging.ZapLogger
}

func New(
	cfg Config,
	tokenAuth *jwtauth.JWTAuth,
	storageService StorageService,
	logger *logging.ZapLogger,
) *Server {
	return &Server{
		srv:            nil,
		cfg:            cfg,
		jwtTokenAuth:   tokenAuth,
		storageService: storageService,
		logger:         logger,
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
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

func (s *Server) createMux() *chi.Mux {
	router := chi.NewRouter()

	// commonMiddleware
	panicRecover := commonMiddleware.NewPanicRecover(s.logger)
	loggerContextMiddleware := commonMiddleware.NewLogging()
	requestDecompression := commonMiddleware.NewRequestDecompressor(s.logger)
	responseCompression := commonMiddleware.NewResponseCompressor(s.logger, gzip.BestSpeed)
	prometheusMiddleware := middleware.NewPrometheus()

	//handlers
	invoiceHandler := handlers.NewInvoice(s.storageService, s.logger)
	authHandler := handlers.NewAuth()

	// router
	router.Use(prometheusMiddleware.CreateHandler)
	router.Use(loggerContextMiddleware.CreateHandler)
	router.Use(panicRecover.CreateHandler)
	router.Route("/api/user/", func(router chi.Router) {
		router.Post("/register", authHandler.Register)
		router.Post("/login", authHandler.Login)
		router.With(
			// jwtauth.Verifier(s.jwtTokenAuth),
			// jwtauth.Authenticator(s.jwtTokenAuth),
			requestDecompression.CreateHandler,
			responseCompression.CreateHandler,
		).Route("/invoice/", func(router chi.Router) {
			router.Post("/create", invoiceHandler.Upload)
			router.Get("/get", invoiceHandler.Get)
		})
	})

	return router
}
