package main

import (
	"context"
	"fmt"
	"go-invoice-service/api-service/cmd/config"
	"go-invoice-service/api-service/internal/http"
	"go-invoice-service/api-service/internal/services"
	"go-invoice-service/common/pkg/jwtfactory"
	"go-invoice-service/common/pkg/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := logging.NewZapLogger(zapcore.DebugLevel)
	if err != nil {
		log.Fatal(err)
	}

	rootCtx, cancelCtx := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT,
	)
	defer cancelCtx()

	storageService, err := services.NewStorage(cfg.StorageConfig)
	if err != nil {
		logger.ErrorCtx(rootCtx, "Failed to create storage service", zap.Error(err))
	}
	defer storageService.Close()

	tokenFactory := jwtfactory.New(cfg.JWTConfig)

	httpServer := http.NewServer(cfg.HTTPServerConfig, tokenFactory.GetJWTAuth(), storageService, logger)

	if err := run(rootCtx, httpServer, logger); err != nil {
		logger.ErrorCtx(rootCtx, "Service shutdown with error", zap.Error(err))
	} else {
		logger.InfoCtx(rootCtx, "Service shutdown gracefully")
	}
}

func run(
	rootCtx context.Context,
	httpServer *http.Server,
	logger *logging.ZapLogger,
) error {
	g, ctx := errgroup.WithContext(rootCtx)

	g.Go(func() error {
		defer logger.InfoCtx(ctx, "HTTP server stopped")

		if err := httpServer.Run(); err != nil {
			return fmt.Errorf("HTTP server error: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("goroutine error occured: %w", err)
	}

	return nil
}
