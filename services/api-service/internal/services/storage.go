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
		Invoice: convertInvoice(invoice),
	}
	_, err := s.storageClient.Upload(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to upload invoice: %w", err)
	}
	s.logger.InfoCtx(ctx, fmt.Sprintf("Invoice %s uploaded successfully", invoice.ID))
	return nil
}

func convertInvoice(invoice dto.Invoice) *types.Invoice {
	return &types.Invoice{
		Id:         convertUUID(invoice.ID),
		CustomerId: convertUUID(invoice.CustomerID),
		Amount:     &invoice.Amount,
		Currency:   &invoice.Currency,
		DueDate:    convertTime(invoice.DueDate),
		CreatedAt:  convertTime(invoice.CreatedAt),
		UpdatedAt:  convertTime(invoice.UpdatedAt),
		Items:      convertItems(invoice.Items),
		Notes:      &invoice.Notes,
	}
}

func convertTime(date time.Time) *timestamppb.Timestamp {
	return &timestamppb.Timestamp{
		Seconds: date.Unix(),
	}
}

func convertUUID(id uuid.UUID) *types.UUID {
	return &types.UUID{
		Value: id[:],
	}
}

func convertItems(items []dto.Item) []*types.Item {
	res := make([]*types.Item, len(items))

	for i, item := range items {
		res[i] = convertItem(item)
	}

	return res
}

func convertItem(item dto.Item) *types.Item {
	return &types.Item{
		Description: &item.Description,
		Quantity:    &item.Quantity,
		UnitPrice:   &item.UnitPrice,
		Total:       &item.Total,
	}
}
