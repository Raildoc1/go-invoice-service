package servers

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	pb "go-invoice-service/common/protocol/proto/apiservice"
	"go-invoice-service/common/protocol/proto/types"
	"storage-service/internal/dto"
)

var _ pb.InvoiceStorageServer = (*InvoiceServer)(nil)

type InvoiceService interface {
	AddNew(ctx context.Context, invoice dto.Invoice) error
}

type InvoiceServer struct {
	pb.UnimplementedInvoiceStorageServer
	service InvoiceService
}

func NewInvoiceServer(service InvoiceService) *InvoiceServer {
	return &InvoiceServer{
		service: service,
	}
}

func (s *InvoiceServer) Upload(ctx context.Context, request *pb.UploadRequest) (*pb.UploadResponse, error) {
	invoice, err := convertInvoice(request)
	if err != nil {
		errMsg := fmt.Sprintf("failed to convert invoice %v", err)
		return &pb.UploadResponse{Error: &errMsg}, nil
	}

	err = s.service.AddNew(ctx, invoice)
	if err != nil {
		errMsg := fmt.Sprintf("failed to add new invoice %v", err)
		return &pb.UploadResponse{Error: &errMsg}, nil
	}

	return &pb.UploadResponse{}, nil
}

func convertInvoice(request *pb.UploadRequest) (dto.Invoice, error) {
	id, err := uuid.FromBytes(request.Invoice.Id.Value)
	if err != nil {
		return dto.Invoice{}, fmt.Errorf("invalid invoice id: %w", err)
	}

	customerId, err := uuid.FromBytes(request.Invoice.CustomerId.Value)
	if err != nil {
		return dto.Invoice{}, fmt.Errorf("invalid customer id: %w", err)
	}

	return dto.Invoice{
		ID:         id,
		CustomerID: customerId,
		Amount:     *request.Invoice.Amount,
		Currency:   *request.Invoice.Currency,
		DueDate:    request.Invoice.DueDate.AsTime(),
		CreatedAt:  request.Invoice.CreatedAt.AsTime(),
		UpdatedAt:  request.Invoice.UpdatedAt.AsTime(),
		Items:      convertItems(request.Invoice.Items),
		Notes:      *request.Invoice.Notes,
	}, nil
}

func convertItems(items []*types.Item) []dto.Item {
	res := make([]dto.Item, len(items))
	for i, item := range items {
		res[i] = convertItem(item)
	}
	return res
}

func convertItem(item *types.Item) dto.Item {
	return dto.Item{
		Description: *item.Description,
		Quantity:    *item.Quantity,
		UnitPrice:   *item.UnitPrice,
		Total:       *item.Total,
	}
}
