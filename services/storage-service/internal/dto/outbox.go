package dto

import "go-invoice-service/common/protocol/kafka"

type OutboxMessageStencil struct {
	Topic   kafka.Topic
	Payload []byte
}

type OutboxMessage struct {
	ID      int64
	Stencil OutboxMessageStencil
}
