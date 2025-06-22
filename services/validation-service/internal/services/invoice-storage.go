package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-invoice-service/common/pkg/logging"
	"go-invoice-service/common/protocol/proto/types"
	pb "go-invoice-service/common/protocol/proto/validation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"validation-service/internal/dto"
)

type StorageConfig struct {
	ServerAddress string
}

type InvoiceStorage struct {
	conn                 *grpc.ClientConn
	invoiceStorageClient pb.InvoiceStorageClient
	logger               *logging.ZapLogger
}

func NewInvoiceStorage(cfg StorageConfig, logger *logging.ZapLogger) (*InvoiceStorage, error) {
	options := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(cfg.ServerAddress, options)
	if err != nil {
		return nil, err
	}
	storageClient := pb.NewInvoiceStorageClient(conn)
	return &InvoiceStorage{
		conn:                 conn,
		invoiceStorageClient: storageClient,
		logger:               logger,
	}, nil
}

func (s *InvoiceStorage) Close() error {
	err := s.conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close gRPC connection: %w", err)
	}
	return nil
}

func (s *InvoiceStorage) GetInvoice(ctx context.Context, id uuid.UUID) (dto.Invoice, dto.InvoiceStatus, error) {
	req := &pb.GetInvoiceRequest{
		Id: convertUUID(id),
	}
	resp, err := s.invoiceStorageClient.Get(ctx, req)
	if err != nil {
		return dto.Invoice{}, dto.NilInvoiceStatus, fmt.Errorf("failed to get invoice: %w", err)
	}
	invoice, err := retrieveInvoice(resp.GetInvoice())
	if err != nil {
		return dto.Invoice{}, dto.NilInvoiceStatus, fmt.Errorf("failed to retrieve invoice: %w", err)
	}
	invoiceStatus, err := retrieveInvoiceStatus(resp.GetStatus())
	if err != nil {
		return dto.Invoice{}, dto.NilInvoiceStatus, fmt.Errorf("failed to retrieve invoice status: %w", err)
	}
	return invoice, invoiceStatus, nil
}

func (s *InvoiceStorage) SetApproved(ctx context.Context, id uuid.UUID) error {
	req := &pb.SetApprovedRequest{
		Id: convertUUID(id),
	}
	_, err := s.invoiceStorageClient.SetApproved(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to set approved: %w", err)
	}
	return nil
}

func (s *InvoiceStorage) SetRejected(ctx context.Context, id uuid.UUID) error {
	req := &pb.SetRejectedRequest{
		Id: convertUUID(id),
	}
	_, err := s.invoiceStorageClient.SetRejected(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to set rejected: %w", err)
	}
	return nil
}

func retrieveInvoiceStatus(status types.InvoiceStatus) (dto.InvoiceStatus, error) {
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

func retrieveInvoice(invoice *types.Invoice) (dto.Invoice, error) {
	id, err := retrieveUUID(invoice.Id)
	if err != nil {
		return dto.Invoice{}, err
	}
	customerId, err := retrieveUUID(invoice.CustomerId)
	if err != nil {
		return dto.Invoice{}, err
	}
	return dto.Invoice{
		ID:         id,
		CustomerID: customerId,
		Amount:     *invoice.Amount,
		Currency:   *invoice.Currency,
		DueDate:    invoice.DueDate.AsTime(),
		CreatedAt:  invoice.CreatedAt.AsTime(),
		UpdatedAt:  invoice.UpdatedAt.AsTime(),
		Items:      retrieveItems(invoice.Items),
		Notes:      *invoice.Notes,
	}, nil
}

func retrieveItems(items []*types.Item) []dto.Item {
	res := make([]dto.Item, len(items))

	for i, item := range items {
		res[i] = retrieveItem(item)
	}

	return res
}

func retrieveItem(item *types.Item) dto.Item {
	return dto.Item{
		Description: *item.Description,
		Quantity:    *item.Quantity,
		UnitPrice:   *item.UnitPrice,
		Total:       *item.Total,
	}
}

func retrieveUUID(id *types.UUID) (uuid.UUID, error) {
	res, err := uuid.FromBytes(id.Value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid: %w", err)
	}
	return res, nil
}

func convertUUID(id uuid.UUID) *types.UUID {
	bytes := [16]byte(id)
	return &types.UUID{
		Value: bytes[:],
	}
}
