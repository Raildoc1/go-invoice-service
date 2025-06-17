package client

import (
	"github.com/shopspring/decimal"
	"time"
)

type InvoiceStatus string

const (
	StatusPending  InvoiceStatus = "PENDING"
	StatusApproved InvoiceStatus = "APPROVED"
	StatusRejected InvoiceStatus = "REJECTED"
)

type Invoice struct {
	ID         string          `json:"id"`
	CustomerID string          `json:"customer_id"`
	Amount     decimal.Decimal `json:"amount"`
	Currency   string          `json:"currency"`
	DueDate    time.Time       `json:"due_date"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	Items      []Item          `json:"items"`
	Notes      string          `json:"notes,omitempty"`
}

type Item struct {
	Description string          `json:"description"`
	Quantity    int32           `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	Total       decimal.Decimal `json:"total"`
}

type UploadInvoiceRequest struct {
	Invoice Invoice `json:"invoice"`
}

type GetInvoiceRequest struct {
	ID string `json:"id"`
}

type GetInvoiceResponse struct {
	Invoice Invoice       `json:"invoice"`
	Status  InvoiceStatus `json:"status"`
}
