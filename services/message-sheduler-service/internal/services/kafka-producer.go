package services

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducer struct {
	producer *kafka.Producer
}

func NewKafkaProducer(serverAddress string) (*KafkaProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": serverAddress,
	})
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{producer: p}, nil
}

func (p *KafkaProducer) Close() {
	p.producer.Close()
}

func (p *KafkaProducer) SendMessage(topic string, payload []byte) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          payload,
	}

	deliveryChan := make(chan kafka.Event)
	err := p.producer.Produce(msg, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka server: %w", err)
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return fmt.Errorf("failed to send message to Kafka server: %w", m.TopicPartition.Error)
	}
	return nil
}
