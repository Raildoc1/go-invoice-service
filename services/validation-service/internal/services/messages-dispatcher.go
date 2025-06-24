package services

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-invoice-service/common/pkg/http/utils"
	"go-invoice-service/common/pkg/logging"
	protocol "go-invoice-service/common/protocol/kafka"
	"go.uber.org/zap"
	"validation-service/internal/dto"
)

type MessagesDispatcherConfig struct {
	PollTimeoutMs int
}

type InvoiceValidator interface {
	Validate(*dto.Invoice) bool
}

type InvoiceStorage interface {
	GetInvoice(ctx context.Context, id uuid.UUID) (*dto.Invoice, dto.InvoiceStatus, error)
	SetApproved(ctx context.Context, id uuid.UUID) error
	SetRejected(ctx context.Context, id uuid.UUID) error
}

type MessageConsumer interface {
	PeekNext(pollTimeoutMs int) ([]byte, error)
	Commit(ctx context.Context) error
	ErrIsNoMessage(error) bool
}

type MessagesDispatcher struct {
	cfg              MessagesDispatcherConfig
	invoiceStorage   InvoiceStorage
	messageConsumer  MessageConsumer
	invoiceValidator InvoiceValidator
	logger           *logging.ZapLogger
}

func NewMessagesDispatcher(
	cfg MessagesDispatcherConfig,
	invoiceStorage InvoiceStorage,
	messageConsumer MessageConsumer,
	invoiceValidator InvoiceValidator,
	logger *logging.ZapLogger,
) *MessagesDispatcher {
	return &MessagesDispatcher{
		cfg:              cfg,
		invoiceStorage:   invoiceStorage,
		messageConsumer:  messageConsumer,
		invoiceValidator: invoiceValidator,
		logger:           logger,
	}
}

func (d *MessagesDispatcher) Run(ctx context.Context) <-chan error {
	errCh := make(chan error)

	go func(ctx context.Context) {
		defer close(errCh)
		for {
			if ctx.Err() != nil {
				return
			}

			err := d.tick(ctx, d.HandleMessage)

			if err != nil && !d.messageConsumer.ErrIsNoMessage(err) {
				errCh <- err
			}
		}
	}(ctx)

	return errCh
}

func (d *MessagesDispatcher) tick(
	ctx context.Context,
	handleMessage func(context.Context, []byte) error,
) error {
	msg, err := d.messageConsumer.PeekNext(d.cfg.PollTimeoutMs)
	if err != nil {
		return err
	}

	err = handleMessage(ctx, msg)
	if err != nil {
		return err
	}

	err = d.messageConsumer.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (d *MessagesDispatcher) HandleMessage(ctx context.Context, msg []byte) error {
	newInvoice, err := utils.DecodeJSON[protocol.NewInvoice](bytes.NewBuffer(msg))
	if err != nil {
		return fmt.Errorf("failed to decode new invoice: %w", err)
	}
	d.logger.InfoCtx(ctx, "reading invoice", zap.String("id", newInvoice.ID.String()))
	invoice, invoiceStatus, err := d.invoiceStorage.GetInvoice(ctx, newInvoice.ID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoiceStatus != dto.PendingInvoiceStatus {
		// invoice was already processed, just skip
		d.logger.InfoCtx(ctx, "duplicated validation skipped", zap.String("id", invoice.ID.String()))
		return nil
	}
	d.logger.InfoCtx(ctx, "validating invoice", zap.String("id", invoice.ID.String()))
	if d.invoiceValidator.Validate(invoice) {
		err := d.invoiceStorage.SetApproved(ctx, newInvoice.ID)
		if err != nil {
			return fmt.Errorf("failed to set approved invoice: %w", err)
		}
		d.logger.InfoCtx(ctx, "invoice approved", zap.String("id", invoice.ID.String()))
	} else {
		err := d.invoiceStorage.SetRejected(ctx, newInvoice.ID)
		if err != nil {
			return fmt.Errorf("failed to set rejected invoice: %w", err)
		}
		d.logger.InfoCtx(ctx, "invoice rejected", zap.String("id", invoice.ID.String()))
	}
	return nil
}
