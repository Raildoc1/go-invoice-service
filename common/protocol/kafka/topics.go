package kafka

import "github.com/google/uuid"

type Topic string

const (
	TopicNewInvoice      Topic = "new_invoice"
	TopicInvoiceApproved Topic = "invoice_approved"
	TopicInvoiceRejected Topic = "invoice_rejected"
)

type NewInvoice struct {
	ID uuid.UUID `json:"id"`
}

type ApprovedInvoice struct {
	ID uuid.UUID `json:"id"`
}

type RejectedInvoice struct {
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
	{
		Topic:             TopicInvoiceApproved,
		PartitionsCount:   6,
		ReplicationFactor: 3,
	},
	{
		Topic:             TopicInvoiceRejected,
		PartitionsCount:   6,
		ReplicationFactor: 3,
	},
}
