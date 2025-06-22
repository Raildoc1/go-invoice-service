package services

import (
	"context"
	"database/sql"
	"storage-service/internal/dto"
	"time"
)

type TransactionsManager interface {
	Do(ctx context.Context, f func(ctx context.Context, tx *sql.Tx) error) error
	DoOpts(ctx context.Context, opts *sql.TxOptions, f func(ctx context.Context, tx *sql.Tx) error) error
}

type OutboxScheduleRepository interface {
	ScheduleMessage(ctx context.Context, tx *sql.Tx, message dto.OutboxMessageStencil, sendAt time.Time) error
}
