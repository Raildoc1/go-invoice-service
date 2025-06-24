package kafka

import (
	"context"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go-invoice-service/common/pkg/logging"
	"go-invoice-service/common/pkg/timeutils"
	"go.uber.org/zap"
	"time"
)

var (
	ErrPartitionEOF = errors.New("PartitionEOF has been reached")
	ErrNoMessage    = errors.New("no messages polled")
)

type ConsumerMetrics interface {
	IncKafkaTotalConsumedMessages(ctx context.Context, topic, consumerGroupID string)
}

type ConsumerConfig struct {
	ServerAddress string
	RetryAttempts []time.Duration
}

type KafkaConsumer struct {
	cfg      ConsumerConfig
	metrics  ConsumerMetrics
	consumer *kafka.Consumer
	groupID  string
	topic    string
	logger   *logging.ZapLogger
}

func NewKafkaConsumer(
	cfg ConsumerConfig,
	metrics ConsumerMetrics,
	groupID string,
	topic string,
	logger *logging.ZapLogger,
) (*KafkaConsumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  cfg.ServerAddress,
		"group.id":           groupID,
		"enable.auto.commit": false,
		"auto.offset.reset":  "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer %w", err)
	}
	err = consumer.Subscribe(topic, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to kafka topic: %w", err)
	}
	return &KafkaConsumer{
		cfg:      cfg,
		metrics:  metrics,
		consumer: consumer,
		groupID:  groupID,
		topic:    topic,
		logger:   logger,
	}, nil
}

func (c *KafkaConsumer) Close() error {
	return c.consumer.Close()
}

func (c *KafkaConsumer) PeekNext(pollTimeoutMs int) ([]byte, error) {
	ev := c.consumer.Poll(pollTimeoutMs)
	switch e := ev.(type) {
	case *kafka.Message:
		return e.Value, nil

	case kafka.PartitionEOF:
		return nil, ErrPartitionEOF

	case kafka.Error:
		return nil, e

	case nil:
		return nil, ErrNoMessage
	}

	return nil, errors.New("unknown kafka event")
}

func (c *KafkaConsumer) Commit(ctx context.Context) error {
	err := timeutils.Retry(
		ctx,
		c.cfg.RetryAttempts,
		func(ctx context.Context) error {
			c.logger.InfoCtx(ctx, "commiting kafka offset")
			_, err := c.consumer.Commit()
			return err
		},
		func(ctx context.Context, err error) bool {
			c.logger.ErrorCtx(ctx, "kafka commit offset fail", zap.Error(err))
			return true
		},
		true,
	)
	c.logger.InfoCtx(ctx, "kafka offset commited")
	if err != nil {
		return fmt.Errorf("failed to commit message %w", err)
	} else {
		c.metrics.IncKafkaTotalConsumedMessages(ctx, c.topic, c.groupID)
	}
	return nil
}

func (c *KafkaConsumer) ErrIsNoMessage(err error) bool {
	return errors.Is(err, ErrNoMessage)
}
