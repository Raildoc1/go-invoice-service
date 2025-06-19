package kafka

import "github.com/google/uuid"

type Topic string

const (
	TopicNewInvoice = "new_invoice"
)

type NewInvoice struct {
	ID uuid.UUID `json:"id"`
}
