package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-invoice-service/api-service/internal/dto"
	"go-invoice-service/common/pkg/logging"
	pb "go-invoice-service/common/protocol/proto/apiservice"
	"go-invoice-service/common/protocol/proto/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type StorageConfig struct {
	ServerAddress string
}

type Storage struct {
	conn          *grpc.ClientConn
	storageClient pb.InvoiceStorageClient
	logger        *logging.ZapLogger
}

func NewStorage(cfg StorageConfig, logger *logging.ZapLogger) (*Storage, error) {
	options := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(cfg.ServerAddress, options)
	if err != nil {
		return nil, err
	}
	storageClient := pb.NewInvoiceStorageClient(conn)
	return &Storage{
		conn:          conn,
		storageClient: storageClient,
		logger:        logger,
	}, nil
}

func (s *Storage) Close() error {
	err := s.conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close gRPC connection: %w", err)
	}
	return nil
}

func (s *Storage) Upload(ctx context.Context, invoice dto.Invoice) error {
	req := &pb.UploadRequest{
		Invoice: invoiceToPB(invoice),
	}
	_, err := s.storageClient.Upload(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to upload invoice: %w", err)
	}
	s.logger.InfoCtx(ctx, fmt.Sprintf("Invoice %s uploaded successfully", invoice.ID))
	return nil
}

func (s *Storage) Get(ctx context.Context, id uuid.UUID) (dto.Invoice, dto.InvoiceStatus, error) {
	req := &pb.GetRequest{
		Id: uuidToPB(id),
	}
	resp, err := s.storageClient.Get(ctx, req)
	if err != nil {
		return dto.Invoice{}, "", fmt.Errorf("failed to get invoice: %w", err)
	}
	invoice, err := invoiceFromPB(resp.Invoice)
	if err != nil {
		return dto.Invoice{}, "", fmt.Errorf("failed to read invoice from pb: %w", err)
	}
	status, err := statusFromPB(resp.Status)
	if err != nil {
		return dto.Invoice{}, "", fmt.Errorf("failed to read invoice status from pb: %w", err)
	}
	return *invoice, status, nil
}

func statusFromPB(status *types.InvoiceStatus) (dto.InvoiceStatus, error) {
	switch *status {
	case types.InvoiceStatus_Pending:
		return dto.StatusPending, nil
	case types.InvoiceStatus_Approved:
		return dto.StatusApproved, nil
	case types.InvoiceStatus_Rejected:
		return dto.StatusRejected, nil
	}
	return "", fmt.Errorf("invalid invoice status: %s", *status)
}

func invoiceFromPB(invoice *types.Invoice) (*dto.Invoice, error) {
	id, err := uuidFromPB(invoice.Id)
	if err != nil {
		return nil, err
	}
	customerID, err := uuidFromPB(invoice.CustomerId)
	if err != nil {
		return nil, err
	}
	return &dto.Invoice{
		ID:         id,
		CustomerID: customerID,
		Amount:     *invoice.Amount,
		Currency:   *invoice.Currency,
		DueDate:    invoice.DueDate.AsTime(),
		CreatedAt:  invoice.CreatedAt.AsTime(),
		UpdatedAt:  invoice.UpdatedAt.AsTime(),
		Items:      itemsFromPB(invoice.Items),
		Notes:      *invoice.Notes,
	}, nil
}

func itemsFromPB(items []*types.Item) []dto.Item {
	res := make([]dto.Item, len(items))

	for i, item := range items {
		res[i] = itemFromPB(item)
	}

	return res
}

func itemFromPB(item *types.Item) dto.Item {
	return dto.Item{
		Description: *item.Description,
		Quantity:    *item.Quantity,
		UnitPrice:   *item.UnitPrice,
		Total:       *item.Total,
	}
}

func uuidFromPB(id *types.UUID) (uuid.UUID, error) {
	res, err := uuid.FromBytes(id.Value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID: %w", err)
	}
	return res, nil
}

func invoiceToPB(invoice dto.Invoice) *types.Invoice {
	return &types.Invoice{
		Id:         uuidToPB(invoice.ID),
		CustomerId: uuidToPB(invoice.CustomerID),
		Amount:     &invoice.Amount,
		Currency:   &invoice.Currency,
		DueDate:    timeToPB(invoice.DueDate),
		CreatedAt:  timeToPB(invoice.CreatedAt),
		UpdatedAt:  timeToPB(invoice.UpdatedAt),
		Items:      itemsToPB(invoice.Items),
		Notes:      &invoice.Notes,
	}
}

func timeToPB(date time.Time) *timestamppb.Timestamp {
	return &timestamppb.Timestamp{
		Seconds: date.Unix(),
	}
}

func uuidToPB(id uuid.UUID) *types.UUID {
	return &types.UUID{
		Value: id[:],
	}
}

func itemsToPB(items []dto.Item) []*types.Item {
	res := make([]*types.Item, len(items))

	for i, item := range items {
		res[i] = itemToPB(item)
	}

	return res
}

func itemToPB(item dto.Item) *types.Item {
	return &types.Item{
		Description: &item.Description,
		Quantity:    &item.Quantity,
		UnitPrice:   &item.UnitPrice,
		Total:       &item.Total,
	}
}
