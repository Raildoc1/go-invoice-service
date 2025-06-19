package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"storage-service/internal/data/postgres/generated/queries"
	"storage-service/internal/dto"
	"time"
)

type Outbox struct {
	qs *queries.Queries
}

func NewOutbox(dbtx queries.DBTX) *Outbox {
	return &Outbox{
		qs: queries.New(dbtx),
	}
}

func (r *Outbox) ScheduleMessage(ctx context.Context, tx *sql.Tx, message dto.OutboxMessage, sendAt time.Time) error {
	qs := r.qs.WithTx(tx)

	err := qs.ScheduleMessage(ctx, convertOutboxMessage(message, sendAt))
	if err != nil {
		return fmt.Errorf("schedule message query failed: %w", err)
	}

	return nil
}

func convertOutboxMessage(message dto.OutboxMessage, sendAt time.Time) queries.ScheduleMessageParams {
	return queries.ScheduleMessageParams{
		Payload:    message.Payload,
		Topic:      string(message.Topic),
		NextSendAt: sendAt,
	}
}
