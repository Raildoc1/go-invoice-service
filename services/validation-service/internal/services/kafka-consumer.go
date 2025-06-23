package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/prometheus/client_golang/prometheus"
	protocol "go-invoice-service/common/protocol/kafka"
	"validation-service/internal/metrics"
)

var (
	ErrPartitionEOF = errors.New("PartitionEOF has been reached")
)

const (
	consumerGroupID = "validation-service"
)

type KafkaConsumerConfig struct {
	ServerAddress string
}

type KafkaConsumer struct {
	cfg      KafkaConsumerConfig
	consumer *kafka.Consumer
}

func NewKafkaConsumer(cfg KafkaConsumerConfig) (*KafkaConsumer, error) {
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
		consumer: c,
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
		err := handleMsg(ctx, e.Value)
		if err != nil {
			return err
		}
		_, err = r.consumer.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit message %w", err)
		} else {
			metrics.KafkaTotalConsumedMessages.With(
				prometheus.Labels{
					"topic":             string(protocol.TopicNewInvoice),
					"consumer-group-id": consumerGroupID,
				},
			).Inc()
		}
		return nil

	case kafka.PartitionEOF:
		return ErrPartitionEOF

	case kafka.Error:
		return e
	}
	return nil
}
