package dto

import "go-invoice-service/common/protocol/kafka"

type OutboxMessage struct {
	Topic   kafka.Topic
	Payload []byte
}
