package controllers

import (
	"context"
	"fmt"
	"go-invoice-service/common/pkg/chutils"
	"go-invoice-service/common/pkg/logging"
	"message-sheduler-service/internal/dto"
	"time"
)

type StorageService interface {
	GetOutboxMessages(ctx context.Context, maxCount int32, retryIn time.Duration) ([]dto.OutboxMessage, error)
	DeleteOutboxMessage(ctx context.Context, id int64) error
}

type KafkaProducer interface {
	SendMessage(ctx context.Context, topic string, payload []byte) error
}

type OutboxDispatcher struct {
	cfg            OutboxDispatcherConfig
	storageService StorageService
	kafkaProducer  KafkaProducer
	logger         *logging.ZapLogger
}

type OutboxDispatcherConfig struct {
	NumWorkers       int32
	RetryIn          time.Duration
	DispatchInterval time.Duration
}

func NewOutboxDispatcher(
	cfg OutboxDispatcherConfig,
	storageService StorageService,
	kafkaProducer KafkaProducer,
	logger *logging.ZapLogger,
) *OutboxDispatcher {
	return &OutboxDispatcher{
		cfg:            cfg,
		storageService: storageService,
		kafkaProducer:  kafkaProducer,
		logger:         logger,
	}
}

func (d *OutboxDispatcher) Run(ctx context.Context) <-chan error {
	errChs := make([]<-chan error, d.cfg.NumWorkers+1)

	const overhead int32 = 1 // making buffer length > numWorkers to prevent workers idling while waiting db response
	genOut, genErr := d.messagesGenerator(ctx, d.cfg.NumWorkers*(overhead+1), d.cfg.RetryIn)
	errChs[0] = genErr

	for i := range d.cfg.NumWorkers {
		errChs[i+1] = d.messagesSender(ctx, genOut)
	}

	return chutils.FanIn(errChs...)
}

func (d *OutboxDispatcher) messagesGenerator(
	ctx context.Context,
	buffCap int32,
	retryIn time.Duration,
) (<-chan dto.OutboxMessage, <-chan error) {
	return chutils.Generator[dto.OutboxMessage](
		ctx,
		buffCap,
		d.cfg.DispatchInterval,
		func(ctx context.Context, buffLen int32) ([]dto.OutboxMessage, error) {
			return d.storageService.GetOutboxMessages(ctx, buffCap-buffLen, retryIn)
		},
	)
}

func (d *OutboxDispatcher) messagesSender(ctx context.Context, in <-chan dto.OutboxMessage) <-chan error {
	errCh := make(chan error)

	go func(ctx context.Context) {
		defer close(errCh)

		for msg := range in {
			if ctx.Err() != nil {
				return
			}

			err := d.kafkaProducer.SendMessage(ctx, msg.Topic, msg.Payload)
			if err != nil {
				errCh <- fmt.Errorf("failed to send message to kafka: %w", err)
				continue
			}
			d.logger.InfoCtx(ctx, fmt.Sprintf("message %v sent to topic %s", msg.ID, msg.Topic))

			err = d.storageService.DeleteOutboxMessage(ctx, msg.ID)
			if err != nil {
				errCh <- fmt.Errorf("failed to delete outbox message: %w", err)
				continue
			}
			d.logger.InfoCtx(ctx, fmt.Sprintf("message %v removed from outbox", msg.ID))
		}
	}(ctx)

	return errCh
}
