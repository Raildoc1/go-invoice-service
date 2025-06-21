package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-invoice-service/api-service/cmd/config"
	"go-invoice-service/api-service/internal/httpserver"
	"go-invoice-service/api-service/internal/services"
	"go-invoice-service/common/pkg/jwtfactory"
	"go-invoice-service/common/pkg/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	cfgJSON, err := json.MarshalIndent(cfg, "", "   ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Config:", string(cfgJSON))

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

	storageService, err := services.NewStorage(cfg.StorageConfig, logger)
	if err != nil {
		logger.ErrorCtx(rootCtx, "Failed to create storage service", zap.Error(err))
	}
	defer storageService.Close()

	tokenFactory := jwtfactory.New(cfg.JWTConfig)

	httpServer := httpserver.New(cfg.HTTPServerConfig, tokenFactory.GetJWTAuth(), storageService, logger)

	if err := run(rootCtx, cfg, httpServer, logger); err != nil {
		logger.ErrorCtx(rootCtx, "Service shutdown with error", zap.Error(err))
	} else {
		logger.InfoCtx(rootCtx, "Service shutdown gracefully")
	}
}

func run(
	rootCtx context.Context,
	cfg *config.Config,
	httpServer *httpserver.Server,
	logger *logging.ZapLogger,
) error {
	g, ctx := errgroup.WithContext(rootCtx)

	context.AfterFunc(ctx, func() {
		timeoutCtx, cancelCtx := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancelCtx()

		<-timeoutCtx.Done()
		log.Fatal("failed to gracefully shutdown the server")
	})

	g.Go(func() error {
		defer logger.InfoCtx(ctx, "HTTP server stopped")

		logger.InfoCtx(ctx, fmt.Sprintf("starting HTTP server '%s'", cfg.HTTPServerConfig.ServerAddress))
		if err := httpServer.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server error: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		defer logger.InfoCtx(ctx, "HTTP server shutdown")
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPServerConfig.ShutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown HTTP server error: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("goroutine error occured: %w", err)
	}

	return nil
}
