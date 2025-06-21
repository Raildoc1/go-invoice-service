package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"go-invoice-service/common/protocol/kafka"
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

func (r *Outbox) ScheduleMessage(ctx context.Context, tx *sql.Tx, message dto.OutboxMessageStencil, sendAt time.Time) error {
	qs := r.qs.WithTx(tx)

	err := qs.ScheduleMessage(ctx, convertOutboxMessage(message, sendAt))
	if err != nil {
		return fmt.Errorf("schedule message query failed: %w", err)
	}

	return nil
}

func (r *Outbox) GetMessages(ctx context.Context, tx *sql.Tx, limit int32) ([]dto.OutboxMessage, error) {
	qs := r.qs.WithTx(tx)

	res, err := qs.GetMessages(ctx, createGetMessagesParams(limit, time.Now().UTC()))
	if err != nil {
		return nil, fmt.Errorf("get outbox messages query failed: %w", err)
	}

	return retrieveMessages(res), nil
}

func (r *Outbox) UpdateMessagesSendTime(
	ctx context.Context,
	tx *sql.Tx,
	ids []int64,
	deltaSec int64,
) error {
	qs := r.qs.WithTx(tx)

	err := qs.IncreaseNextSendAt(ctx, createIncreaseNextSendAtParams(ids, deltaSec))
	if err != nil {
		return fmt.Errorf("update messages send time query failed: %w", err)
	}

	return nil
}

func (r *Outbox) Delete(ctx context.Context, tx *sql.Tx, id int64) error {
	qs := r.qs.WithTx(tx)

	err := qs.DeleteMessage(ctx, id)
	if err != nil {
		return fmt.Errorf("delete message query failed: %w", err)
	}

	return nil
}

func createIncreaseNextSendAtParams(ids []int64, deltaSec int64) queries.IncreaseNextSendAtParams {
	return queries.IncreaseNextSendAtParams{
		TimeToAddSec: deltaSec,
		Ids:          ids,
	}
}

func retrieveMessages(messages []queries.GetMessagesRow) []dto.OutboxMessage {
	res := make([]dto.OutboxMessage, len(messages))

	for i, m := range messages {
		res[i] = retrieveMessage(m)
	}

	return res
}

func retrieveMessage(m queries.GetMessagesRow) dto.OutboxMessage {
	return dto.OutboxMessage{
		ID: m.ID,
		Stencil: dto.OutboxMessageStencil{
			Topic:   kafka.Topic(m.Topic),
			Payload: m.Payload,
		},
	}
}

func createGetMessagesParams(limit int32, now time.Time) queries.GetMessagesParams {
	return queries.GetMessagesParams{
		NextSendAt: now,
		Limit:      limit,
	}
}

func convertOutboxMessage(message dto.OutboxMessageStencil, sendAt time.Time) queries.ScheduleMessageParams {
	return queries.ScheduleMessageParams{
		Payload:    message.Payload,
		Topic:      string(message.Topic),
		NextSendAt: sendAt,
	}
}
