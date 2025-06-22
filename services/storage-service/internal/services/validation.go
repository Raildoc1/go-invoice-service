package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"go-invoice-service/common/protocol/kafka"
	"storage-service/internal/dto"
	"time"
)

type InvoiceRepository interface {
	GetInvoice(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*dto.Invoice, dto.InvoiceStatus, error)
	SetStatus(ctx context.Context, tx *sql.Tx, id uuid.UUID, status dto.InvoiceStatus) error
}

type Validation struct {
	tm         TransactionsManager
	invoiceRep InvoiceRepository
	outboxRep  OutboxScheduleRepository
}

func NewValidation(
	tm TransactionsManager,
	invoiceRep InvoiceRepository,
	outboxRep OutboxScheduleRepository,
) *Validation {
	return &Validation{
		tm:         tm,
		invoiceRep: invoiceRep,
		outboxRep:  outboxRep,
	}
}

func (s *Validation) Get(ctx context.Context, id uuid.UUID) (*dto.Invoice, dto.InvoiceStatus, error) {

	var resInvoice *dto.Invoice
	var resStatus dto.InvoiceStatus

	err := s.tm.DoOpts(ctx,
		&sql.TxOptions{
			Isolation: sql.LevelSerializable,
			ReadOnly:  true,
		},
		func(ctx context.Context, tx *sql.Tx) error {
			invoice, status, err := s.invoiceRep.GetInvoice(ctx, tx, id)
			if err != nil {
				return err
			}
			resInvoice = invoice
			resStatus = status
			return nil
		},
	)

	if err != nil {
		return nil, dto.StatusNil, err
	}

	return resInvoice, resStatus, nil
}

func (s *Validation) SetApproved(ctx context.Context, id uuid.UUID) error {
	return s.tm.Do(ctx, func(ctx context.Context, tx *sql.Tx) error {
		err := s.invoiceRep.SetStatus(ctx, tx, id, dto.StatusApproved)
		if err != nil {
			return fmt.Errorf("failed to set approved status: %w", err)
		}

		payload := kafka.ApprovedInvoice{
			ID: id,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshalling new invoice kafka message failed: %w", err)
		}
		msg := dto.OutboxMessageStencil{
			Topic:   kafka.TopicInvoiceApproved,
			Payload: payloadJSON,
		}

		err = s.outboxRep.ScheduleMessage(ctx, tx, msg, time.Now().UTC())
		if err != nil {
			return fmt.Errorf("failed to write message to outbox: %w", err)
		}

		return nil
	})
}

func (s *Validation) SetRejected(ctx context.Context, id uuid.UUID) error {
	return s.tm.Do(ctx, func(ctx context.Context, tx *sql.Tx) error {
		err := s.invoiceRep.SetStatus(ctx, tx, id, dto.StatusRejected)
		if err != nil {
			return fmt.Errorf("failed to set rejected status: %w", err)
		}

		payload := kafka.RejectedInvoice{
			ID: id,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshalling new invoice kafka message failed: %w", err)
		}
		msg := dto.OutboxMessageStencil{
			Topic:   kafka.TopicInvoiceRejected,
			Payload: payloadJSON,
		}

		err = s.outboxRep.ScheduleMessage(ctx, tx, msg, time.Now().UTC())
		if err != nil {
			return fmt.Errorf("failed to write message to outbox: %w", err)
		}

		return nil
	})
}
