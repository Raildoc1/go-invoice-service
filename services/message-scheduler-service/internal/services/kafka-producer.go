package services

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaMetrics interface {
	IncKafkaTotalProduceMessages(ctx context.Context, topic string)
	IncKafkaTotalProducedBytes(ctx context.Context, topic string, bytesCount int64)
}

type KafkaProducerConfig struct {
	ServerAddress string
}

type KafkaProducer struct {
	producer *kafka.Producer
	metrics  KafkaMetrics
}

func NewKafkaProducer(cfg KafkaProducerConfig, metrics KafkaMetrics) (*KafkaProducer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.ServerAddress,
	})
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{
		producer: producer,
		metrics:  metrics,
	}, nil
}

func (p *KafkaProducer) Close() {
	p.producer.Close()
}

func (p *KafkaProducer) SendMessage(ctx context.Context, topic string, payload []byte) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          payload,
	}

	deliveryChan := make(chan kafka.Event)
	err := p.producer.Produce(msg, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka server: %w", err)
	}
	p.metrics.IncKafkaTotalProduceMessages(ctx, topic)
	p.metrics.IncKafkaTotalProducedBytes(ctx, topic, int64(len(payload)))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-deliveryChan:
			m := e.(*kafka.Message)
			if m.TopicPartition.Error != nil {
				return fmt.Errorf("failed to send message to Kafka server: %w", m.TopicPartition.Error)
			}
			return nil
		}
	}
}
