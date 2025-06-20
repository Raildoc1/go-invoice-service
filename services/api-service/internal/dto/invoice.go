package dto

import (
	"github.com/google/uuid"
	"time"
)

type Invoice struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Amount     int64
	Currency   string
	DueDate    time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Items      []Item
	Notes      string
}

type Item struct {
	Description string
	Quantity    int32
	UnitPrice   int64
	Total       int64
}
