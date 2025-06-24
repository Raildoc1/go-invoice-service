package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-invoice-service/common/pkg/logging"
	"go-invoice-service/common/pkg/meterutils"
	kafkaProtocol "go-invoice-service/common/protocol/kafka"
	storagepb "go-invoice-service/common/protocol/proto/validation"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"validation-service/cmd/config"
	"validation-service/internal/kafka"
	"validation-service/internal/metrics"
	"validation-service/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	mp, err := meterutils.SetupMeterProvider(cfg.OpenTelemetryConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer mp.Shutdown(context.Background())

	metricsCollector := metrics.MustInitCustomMetric()

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

	kafkaConsumer, err := kafka.NewKafkaConsumer(
		cfg.KafkaConsumerConfig,
		metricsCollector,
		"validation-service",
		string(kafkaProtocol.TopicNewInvoice),
		logger,
	)
	if err != nil {
		logger.FatalCtx(rootCtx, "Failed to create consumer", zap.Error(err))
	}
	defer kafkaConsumer.Close()

	options := grpc.WithTransportCredentials(insecure.NewCredentials())
	storageServiceConnection, err := grpc.NewClient(cfg.StorageAddress, options)
	if err != nil {
		logger.FatalCtx(rootCtx, "Failed to connect to storage service", zap.Error(err))
	}
	defer storageServiceConnection.Close()

	invoiceStorageClient := storagepb.NewInvoiceStorageClient(storageServiceConnection)
	storageService := services.NewInvoiceStorage(invoiceStorageClient, metricsCollector, logger)
	validationService := services.NewValidator()

	messagesDispatcher := services.NewMessagesDispatcher(
		cfg.KafkaDispatcherConfig,
		storageService,
		kafkaConsumer,
		validationService,
		logger,
	)

	if err := run(rootCtx, cfg, messagesDispatcher, logger); err != nil {
		logger.ErrorCtx(rootCtx, "Service shutdown with error", zap.Error(err))
	} else {
		logger.InfoCtx(rootCtx, "Service shutdown gracefully")
	}
}

func run(
	rootCtx context.Context,
	cfg *config.Config,
	messagesDispatcher *services.MessagesDispatcher,
	logger *logging.ZapLogger,
) error {
	g, ctx := errgroup.WithContext(rootCtx)

	g.Go(func() error {
		defer logger.InfoCtx(ctx, "Kafka Dispatching Finished")
		errCh := messagesDispatcher.Run(ctx)
		for err := range errCh {
			logger.ErrorCtx(ctx, "Kafka Dispatching Error", zap.Error(err))
		}
		return nil
	})

	promServer := meterutils.NewPrometheusServer(cfg.PrometheusConfig)

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
