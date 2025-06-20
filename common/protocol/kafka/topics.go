package kafka

import "github.com/google/uuid"

type Topic string

const (
	TopicNewInvoice Topic = "new_invoice"
)

type NewInvoice struct {
	ID uuid.UUID `json:"id"`
}

type TopicSettings struct {
	Topic             Topic
	PartitionsCount   int
	ReplicationFactor int
}

var Topics = []TopicSettings{
	{
		Topic:             TopicNewInvoice,
		PartitionsCount:   6,
		ReplicationFactor: 3,
	},
}
