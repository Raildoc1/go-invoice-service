package main

import (
	"context"
	"fmt"
	"go-invoice-service/common/pkg/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"log"
	"message-sheduler-service/cmd/config"
	"message-sheduler-service/internal/controllers"
	"message-sheduler-service/internal/services"
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

	kafkaProducer, err := services.NewKafkaProducer(cfg.KafkaProducerConfig)
	if err != nil {
		logger.ErrorCtx(rootCtx, "Failed to create kafka producer", zap.Error(err))
	}
	defer kafkaProducer.Close()

	storageService, err := services.NewStorage(cfg.StorageConfig)
	if err != nil {
		logger.ErrorCtx(rootCtx, "Failed to create storage service", zap.Error(err))
	}
	defer storageService.Close()

	outboxDispatcher := controllers.NewOutboxDispatcher(cfg.OutboxDispatcherConfig, storageService, kafkaProducer)

	if err := run(rootCtx, outboxDispatcher, logger); err != nil {
		logger.ErrorCtx(rootCtx, "Service shutdown with error", zap.Error(err))
	} else {
		logger.InfoCtx(rootCtx, "Service shutdown gracefully")
	}
}

func run(
	rootCtx context.Context,
	outboxDispatcher *controllers.OutboxDispatcher,
	logger *logging.ZapLogger,
) error {
	g, ctx := errgroup.WithContext(rootCtx)

	g.Go(func() error {
		defer logger.InfoCtx(ctx, "Outbox Dispatching Finished")
		errCh := outboxDispatcher.Run(ctx)
		for err := range errCh {
			logger.ErrorCtx(ctx, "Outbox Dispatching Error", zap.Error(err))
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("goroutine error occured: %w", err)
	}

	return nil
}
