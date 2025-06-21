package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-invoice-service/common/pkg/logging"
	"go-invoice-service/common/pkg/transactions"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"log"
	"os/signal"
	"storage-service/cmd/config"
	"storage-service/internal/data/postgres"
	"storage-service/internal/data/repositories"
	"storage-service/internal/grpc"
	"storage-service/internal/services"
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

	db, err := postgres.Create(cfg.PostgresConfig)
	if err != nil {
		log.Fatal(err)
	}

	tm := transactions.NewManager(db)

	invoiceRepository := repositories.NewInvoice(db)
	outboxRepository := repositories.NewOutbox(db)

	invoiceService := services.NewInvoice(tm, invoiceRepository, outboxRepository)
	outboxService := services.NewOutbox(tm)

	grpcServer := grpc.NewServer(cfg.GRPCConfig, invoiceService, outboxService)

	if err := run(rootCtx, grpcServer, logger); err != nil {
		logger.ErrorCtx(rootCtx, "Service shutdown with error", zap.Error(err))
	} else {
		logger.InfoCtx(rootCtx, "Service shutdown gracefully")
	}
}

func run(rootCtx context.Context, grpcServer *grpc.Server, logger *logging.ZapLogger) error {
	g, ctx := errgroup.WithContext(rootCtx)

	g.Go(func() error {
		defer logger.InfoCtx(ctx, "gRPC server shutdown")
		if err := grpcServer.Run(); err != nil {
			return fmt.Errorf("grpc server error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		defer logger.InfoCtx(ctx, "Shutting down gRPC server")
		<-ctx.Done()
		grpcServer.Shutdown()
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("goroutine error occured: %w", err)
	}

	return nil
}
