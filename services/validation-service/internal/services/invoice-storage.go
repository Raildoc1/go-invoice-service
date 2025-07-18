package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-invoice-service/common/pkg/logging"
	"go-invoice-service/common/protocol/proto/types"
	pb "go-invoice-service/common/protocol/proto/validation"
	"validation-service/internal/dto"
)

type InvoiceStorageClient interface {
	pb.InvoiceStorageClient
}

type InvoicesMetrics interface {
	IncTotalHandledInvoices(ctx context.Context, status string)
}

type InvoiceStorageService struct {
	invoiceStorageClient InvoiceStorageClient
	metrics              InvoicesMetrics
	logger               *logging.ZapLogger
}

func NewInvoiceStorage(
	invoiceStorageClient InvoiceStorageClient,
	metrics InvoicesMetrics,
	logger *logging.ZapLogger,
) *InvoiceStorageService {
	return &InvoiceStorageService{
		invoiceStorageClient: invoiceStorageClient,
		metrics:              metrics,
		logger:               logger,
	}
}

func (s *InvoiceStorageService) GetInvoice(ctx context.Context, id uuid.UUID) (*dto.Invoice, dto.InvoiceStatus, error) {
	req := &pb.GetInvoiceRequest{
		Id: uuidToProto(id),
	}
	resp, err := s.invoiceStorageClient.Get(ctx, req)
	if err != nil {
		return nil, dto.NilInvoiceStatus, fmt.Errorf("failed to get invoice: %w", err)
	}
	invoice, err := invoiceFromProto(resp.GetInvoice())
	if err != nil {
		return nil, dto.NilInvoiceStatus, fmt.Errorf("failed to retrieve invoice: %w", err)
	}
	invoiceStatus, err := invoiceStatusFromProto(resp.GetStatus())
	if err != nil {
		return nil, dto.NilInvoiceStatus, fmt.Errorf("failed to retrieve invoice status: %w", err)
	}
	return invoice, invoiceStatus, nil
}

func (s *InvoiceStorageService) SetApproved(ctx context.Context, id uuid.UUID) error {
	req := &pb.SetApprovedRequest{
		Id: uuidToProto(id),
	}
	_, err := s.invoiceStorageClient.SetApproved(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to set approved: %w", err)
	}
	s.metrics.IncTotalHandledInvoices(ctx, "approved")
	return nil
}

func (s *InvoiceStorageService) SetRejected(ctx context.Context, id uuid.UUID) error {
	req := &pb.SetRejectedRequest{
		Id: uuidToProto(id),
	}
	_, err := s.invoiceStorageClient.SetRejected(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to set rejected: %w", err)
	}
	s.metrics.IncTotalHandledInvoices(ctx, "rejected")
	return nil
}

func invoiceStatusFromProto(status types.InvoiceStatus) (dto.InvoiceStatus, error) {
	switch status {
	case types.InvoiceStatus_Pending:
		return dto.PendingInvoiceStatus, nil
	case types.InvoiceStatus_Approved:
		return dto.ApprovedInvoiceStatus, nil
	case types.InvoiceStatus_Rejected:
		return dto.RejectedInvoiceStatus, nil
	}
	return dto.NilInvoiceStatus, fmt.Errorf("invalid invoice status %s", status.String())
}

func invoiceFromProto(invoice *types.Invoice) (*dto.Invoice, error) {
	id, err := uuidFromProto(invoice.Id)
	if err != nil {
		return nil, err
	}
	customerId, err := uuidFromProto(invoice.CustomerId)
	if err != nil {
		return nil, err
	}
	return &dto.Invoice{
		ID:         id,
		CustomerID: customerId,
		Amount:     *invoice.Amount,
		Currency:   *invoice.Currency,
		DueDate:    invoice.DueDate.AsTime(),
		CreatedAt:  invoice.CreatedAt.AsTime(),
		UpdatedAt:  invoice.UpdatedAt.AsTime(),
		Items:      itemsFromProto(invoice.Items),
		Notes:      *invoice.Notes,
	}, nil
}

func itemsFromProto(items []*types.Item) []dto.Item {
	res := make([]dto.Item, len(items))

	for i, item := range items {
		res[i] = timeFromProto(item)
	}

	return res
}

func timeFromProto(item *types.Item) dto.Item {
	return dto.Item{
		Description: *item.Description,
		Quantity:    *item.Quantity,
		UnitPrice:   *item.UnitPrice,
		Total:       *item.Total,
	}
}

func uuidFromProto(id *types.UUID) (uuid.UUID, error) {
	res, err := uuid.FromBytes(id.Value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid: %w", err)
	}
	return res, nil
}

func uuidToProto(id uuid.UUID) *types.UUID {
	return &types.UUID{
		Value: id[:],
	}
}
