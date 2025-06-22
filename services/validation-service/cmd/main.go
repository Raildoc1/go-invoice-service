package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-invoice-service/common/pkg/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"log"
	"os/signal"
	"syscall"
	"validation-service/cmd/config"
	"validation-service/internal/controllers"
	"validation-service/internal/services"
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

	kafkaConsumer, err := services.NewKafkaConsumer(cfg.KafkaConsumerConfig)
	if err != nil {
		logger.ErrorCtx(rootCtx, "Failed to create kafka producer", zap.Error(err))
	}
	defer kafkaConsumer.Close()

	storageService, err := services.NewInvoiceStorage(cfg.StorageConfig, logger)
	if err != nil {
		logger.ErrorCtx(rootCtx, "Failed to create storage service", zap.Error(err))
	}
	defer storageService.Close()

	kafkaDispatcher := controllers.NewKafkaDispatcher(
		cfg.KafkaDispatcherConfig,
		kafkaConsumer,
		storageService,
		logger,
	)

	if err := run(rootCtx, kafkaDispatcher, logger); err != nil {
		logger.ErrorCtx(rootCtx, "Service shutdown with error", zap.Error(err))
	} else {
		logger.InfoCtx(rootCtx, "Service shutdown gracefully")
	}
}

func run(rootCtx context.Context, outboxDispatcher *controllers.KafkaDispatcher, logger *logging.ZapLogger) error {
	g, ctx := errgroup.WithContext(rootCtx)

	g.Go(func() error {
		defer logger.InfoCtx(ctx, "Kafka Dispatching Finished")
		errCh := outboxDispatcher.Run(ctx)
		for err := range errCh {
			logger.ErrorCtx(ctx, "Kafka Dispatching Error", zap.Error(err))
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("goroutine error occured: %w", err)
	}

	return nil
}
