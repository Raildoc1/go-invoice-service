package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"go-invoice-service/common/protocol/kafka"
	"storage-service/internal/dto"
	"time"
)

type InvoiceAddRepository interface {
	Add(ctx context.Context, tx *sql.Tx, invoice dto.Invoice, status dto.InvoiceStatus) error
}

type Invoice struct {
	tm         TransactionsManager
	invoiceRep InvoiceAddRepository
	outboxRep  OutboxScheduleRepository
}

func NewInvoice(
	tm TransactionsManager,
	invoiceRep InvoiceAddRepository,
	outboxRep OutboxScheduleRepository,
) *Invoice {
	return &Invoice{
		tm:         tm,
		invoiceRep: invoiceRep,
		outboxRep:  outboxRep,
	}
}

func (s *Invoice) AddNew(ctx context.Context, invoice dto.Invoice) error {
	return s.tm.Do(ctx, func(ctx context.Context, tx *sql.Tx) error {
		err := s.invoiceRep.Add(ctx, tx, invoice, dto.StatusPending)
		if err != nil {
			return fmt.Errorf("adding invoice failed: %w", err)
		}

		payload := kafka.NewInvoice{
			ID: invoice.ID,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshalling new invoice kafka message failed: %w", err)
		}
		msg := dto.OutboxMessageStencil{
			Topic:   kafka.TopicNewInvoice,
			Payload: payloadJSON,
		}
		err = s.outboxRep.ScheduleMessage(ctx, tx, msg, time.Now().UTC())
		if err != nil {
			return fmt.Errorf("scheduled message failed: %w", err)
		}

		return nil
	})
}
