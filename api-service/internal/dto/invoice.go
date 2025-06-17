package dto

import (
	"github.com/shopspring/decimal"
	"time"
)

type Invoice struct {
	ID         string
	CustomerID string
	Amount     decimal.Decimal
	Currency   string
	DueDate    time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Items      []Item
	Notes      string
}

type Item struct {
	Description string
	Quantity    int
	UnitPrice   decimal.Decimal
	Total       decimal.Decimal
}
