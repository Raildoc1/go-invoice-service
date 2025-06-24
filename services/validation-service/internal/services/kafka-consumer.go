package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go-invoice-service/common/pkg/logging"
	"go-invoice-service/common/pkg/timeutils"
	protocol "go-invoice-service/common/protocol/kafka"
	"go.uber.org/zap"
	"time"
)

var (
	ErrPartitionEOF = errors.New("PartitionEOF has been reached")
)

const (
	consumerGroupID = "validation-service"
)

type KafkaConsumerMetrics interface {
	IncKafkaTotalConsumedMessages(ctx context.Context, topic, consumerGroupID string)
}

type KafkaConsumerConfig struct {
	ServerAddress string
	RetryAttempts []time.Duration
}

type KafkaConsumer struct {
	cfg      KafkaConsumerConfig
	metrics  KafkaConsumerMetrics
	consumer *kafka.Consumer
	logger   *logging.ZapLogger
}

func NewKafkaConsumer(
	cfg KafkaConsumerConfig,
	metrics KafkaConsumerMetrics,
	logger *logging.ZapLogger,
) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  cfg.ServerAddress,
		"group.id":           consumerGroupID,
		"enable.auto.commit": false,
		"auto.offset.reset":  "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer %w", err)
	}
	err = c.Subscribe(string(protocol.TopicNewInvoice), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to kafka topic: %w", err)
	}
	return &KafkaConsumer{
		cfg:      cfg,
		metrics:  metrics,
		consumer: c,
		logger:   logger,
	}, nil
}

func (r *KafkaConsumer) Close() error {
	err := r.consumer.Close()
	if err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}
	return nil
}

func (r *KafkaConsumer) HandleNext(
	ctx context.Context,
	pollTimeoutMs int,
	handleMsg func(context.Context, []byte) error,
) error {
	ev := r.consumer.Poll(pollTimeoutMs)
	switch e := ev.(type) {
	case *kafka.Message:
		r.logger.InfoCtx(ctx, "handling kafka message")
		err := handleMsg(ctx, e.Value)
		if err != nil {
			return err
		}
		err = timeutils.Retry(
			ctx,
			r.cfg.RetryAttempts,
			func(ctx context.Context) error {
				r.logger.InfoCtx(ctx, "commiting kafka offset")
				_, err = r.consumer.Commit()
				return err
			},
			func(ctx context.Context, err error) bool {
				r.logger.ErrorCtx(ctx, "kafka commit offset fail", zap.Error(err))
				return true
			},
			true,
		)
		r.logger.InfoCtx(ctx, "kafka offset commited")
		if err != nil {
			return fmt.Errorf("failed to commit message %w", err)
		} else {
			r.metrics.IncKafkaTotalConsumedMessages(ctx, string(protocol.TopicNewInvoice), consumerGroupID)
		}
		return nil

	case kafka.PartitionEOF:
		return ErrPartitionEOF

	case kafka.Error:
		return e
	}
	return nil
}
