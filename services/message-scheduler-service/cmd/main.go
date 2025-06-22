package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-invoice-service/common/pkg/logging"
	"go-invoice-service/common/pkg/promutils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"log"
	"message-sheduler-service/cmd/config"
	"message-sheduler-service/internal/controllers"
	"message-sheduler-service/internal/metrics"
	"message-sheduler-service/internal/services"
	"message-sheduler-service/internal/setup"
	"net/http"
	"os/signal"
	"syscall"
)

func main() {
	metrics.MustInit()

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

	outboxDispatcher := controllers.NewOutboxDispatcher(
		cfg.OutboxDispatcherConfig,
		storageService,
		kafkaProducer,
		logger,
	)

	if err := run(rootCtx, cfg, outboxDispatcher, logger); err != nil {
		logger.ErrorCtx(rootCtx, "Service shutdown with error", zap.Error(err))
	} else {
		logger.InfoCtx(rootCtx, "Service shutdown gracefully")
	}
}

func run(
	rootCtx context.Context,
	cfg *config.Config,
	outboxDispatcher *controllers.OutboxDispatcher,
	logger *logging.ZapLogger,
) error {
	err := setup.EnsureKafkaTopics(rootCtx, cfg.KafkaProducerConfig.ServerAddress, logger)
	if err != nil {
		return fmt.Errorf("failed to setup kafka topics: %w", err)
	}

	g, ctx := errgroup.WithContext(rootCtx)

	g.Go(func() error {
		defer logger.InfoCtx(ctx, "Outbox Dispatching Finished")
		errCh := outboxDispatcher.Run(ctx)
		for err := range errCh {
			logger.ErrorCtx(ctx, "Outbox Dispatching Error", zap.Error(err))
		}
		return nil
	})

	promServer := promutils.NewServer(cfg.PrometheusConfig)

	g.Go(func() error {
		defer logger.InfoCtx(ctx, "Prometheus HTTP server stopped")
		if err := promServer.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		defer logger.InfoCtx(ctx, "Prometheus HTTP server shutdown")
		<-ctx.Done()
		if err := promServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown HTTP server error: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("goroutine error occured: %w", err)
	}

	return nil
}
