package services

import (
	"context"
	"errors"
	"storage-service/internal/dto"
	"time"
)

type Outbox struct {
	tm TransactionsManager
}

func NewOutbox(tm TransactionsManager) *Outbox {
	return &Outbox{
		tm: tm,
	}
}

func (s *Outbox) Get(ctx context.Context, maxCount int32, retryAfter time.Duration) ([]dto.OutboxMessage, error) {
	return nil, errors.New("not implemented")
}

func (s *Outbox) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}
