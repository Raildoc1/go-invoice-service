package controllers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-invoice-service/common/pkg/http/utils"
	"go-invoice-service/common/pkg/logging"
	protocol "go-invoice-service/common/protocol/kafka"
	"go.uber.org/zap"
	"math/rand/v2"
	"time"
	"validation-service/internal/dto"
)

type KafkaConsumer interface {
	HandleNext(ctx context.Context, pollTimeoutMs int, handleMsg func(context.Context, []byte) error) error
}

type InvoiceStorage interface {
	GetInvoice(ctx context.Context, id uuid.UUID) (dto.Invoice, dto.InvoiceStatus, error)
	SetApproved(ctx context.Context, id uuid.UUID) error
	SetRejected(ctx context.Context, id uuid.UUID) error
}

type KafkaDispatcherConfig struct {
	PollTimeoutMs int
}

type KafkaDispatcher struct {
	cfg            KafkaDispatcherConfig
	consumer       KafkaConsumer
	invoiceStorage InvoiceStorage
	logger         *logging.ZapLogger
}

func NewKafkaDispatcher(
	cfg KafkaDispatcherConfig,
	consumer KafkaConsumer,
	invoiceStorage InvoiceStorage,
	logger *logging.ZapLogger,
) *KafkaDispatcher {
	return &KafkaDispatcher{
		cfg:            cfg,
		consumer:       consumer,
		invoiceStorage: invoiceStorage,
		logger:         logger,
	}
}

func (d *KafkaDispatcher) Run(ctx context.Context) <-chan error {
	errCh := make(chan error)

	go func(ctx context.Context) {
		defer close(errCh)
		for {
			if ctx.Err() != nil {
				return
			}
			err := d.consumer.HandleNext(ctx, d.cfg.PollTimeoutMs, d.HandleMessage)
			if err != nil {
				errCh <- err
			}
		}
	}(ctx)

	return errCh
}

func (d *KafkaDispatcher) HandleMessage(ctx context.Context, msg []byte) error {
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
	if d.validate(invoice) {
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

func (d *KafkaDispatcher) validate(_ dto.Invoice) bool {
	time.Sleep(time.Duration(rand.Int()%5_000) * time.Millisecond) // some validation logic
	const approveProbability float64 = 0.9
	return rand.Float64() < approveProbability
}
