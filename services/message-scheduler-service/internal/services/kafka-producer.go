package services

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/prometheus/client_golang/prometheus"
	"message-sheduler-service/internal/metrics"
)

type KafkaProducerConfig struct {
	ServerAddress string
}

type KafkaProducer struct {
	producer *kafka.Producer
}

func NewKafkaProducer(cfg KafkaProducerConfig) (*KafkaProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.ServerAddress,
	})
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{producer: p}, nil
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
	IncKafkaTotalProduceAttempts(topic, err == nil)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka server: %w", err)
	}
	IntTotalKafkaBytesSent(topic, len(payload))

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

func IncKafkaTotalProduceAttempts(topic string, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}

	labels := prometheus.Labels{
		"topic":  topic,
		"status": status,
	}

	metrics.KafkaTotalProduceAttempts.With(labels).Inc()
}

func IntTotalKafkaBytesSent(topic string, bytes int) {
	labels := prometheus.Labels{
		"topic": topic,
	}

	metrics.KafkaTotalProducedBytes.With(labels).Add(float64(bytes))
}
