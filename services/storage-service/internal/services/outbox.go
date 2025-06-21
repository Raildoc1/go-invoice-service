package services

import (
	"context"
	"database/sql"
	"fmt"
	"go-invoice-service/common/pkg/logging"
	"storage-service/internal/dto"
	"time"
)

type OutboxRepository interface {
	GetMessages(ctx context.Context, tx *sql.Tx, limit int32) ([]dto.OutboxMessage, error)
	UpdateMessagesSendTime(ctx context.Context, tx *sql.Tx, ids []int64, deltaSec int64) error
	Delete(ctx context.Context, tx *sql.Tx, id int64) error
}

type Outbox struct {
	tm               TransactionsManager
	outboxRepository OutboxRepository
	logger           *logging.ZapLogger
}

func NewOutbox(tm TransactionsManager, outboxRepository OutboxRepository, logger *logging.ZapLogger) *Outbox {
	return &Outbox{
		tm:               tm,
		outboxRepository: outboxRepository,
		logger:           logger,
	}
}

func (s *Outbox) Get(ctx context.Context, maxCount int32, retryAfter time.Duration) ([]dto.OutboxMessage, error) {
	var res []dto.OutboxMessage
	if err := s.tm.Do(ctx, func(ctx context.Context, tx *sql.Tx) error {
		messages, err := s.outboxRepository.GetMessages(ctx, tx, maxCount)
		if err != nil {
			return fmt.Errorf("failed to get outbox messages: %w", err)
		}
		s.logger.InfoCtx(ctx, fmt.Sprintf("Got %d outbox messages", len(messages)))
		if len(messages) == 0 {
			res = make([]dto.OutboxMessage, 0)
			return nil
		}
		ids := make([]int64, len(messages))
		for i, msg := range messages {
			ids[i] = msg.ID
		}
		err = s.outboxRepository.UpdateMessagesSendTime(ctx, tx, ids, int64(retryAfter.Seconds()))
		if err != nil {
			return fmt.Errorf("failed to update outbox messages' retry time¬: %w", err)
		}
		res = messages
		return nil
	}); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Outbox) Delete(ctx context.Context, id int64) error {
	return s.tm.Do(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := s.outboxRepository.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("failed to delete outbox: %w", err)
		}
		return nil
	})
}
