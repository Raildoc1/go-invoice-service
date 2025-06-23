package servers

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	pb "go-invoice-service/common/protocol/proto/apiservice"
	"go-invoice-service/common/protocol/proto/types"
	"google.golang.org/protobuf/types/known/emptypb"
	"storage-service/internal/dto"
)

var _ pb.InvoiceStorageServer = (*InvoiceServer)(nil)

type InvoiceService interface {
	AddNew(ctx context.Context, invoice *dto.Invoice) error
	Get(ctx context.Context, id uuid.UUID) (*dto.Invoice, dto.InvoiceStatus, error)
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

func (s *InvoiceServer) Upload(ctx context.Context, request *pb.UploadRequest) (*emptypb.Empty, error) {
	invoice, err := invoiceToPB(request)
	if err != nil {
		return nil, fmt.Errorf("failed to convert invoice %w", err)
	}

	err = s.service.AddNew(ctx, invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to add new invoice %w", err)
	}

	return nil, nil
}

func (s *InvoiceServer) Get(ctx context.Context, request *pb.GetRequest) (*pb.GetResponse, error) {
	id, err := uuid.FromBytes(request.Id.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice id: %w", err)
	}

	invoice, status, err := s.service.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice %w", err)
	}

	statusPB, err := statusToProto(status)
	if err != nil {
		return nil, fmt.Errorf("failed to convert invoice status %w", err)
	}

	return &pb.GetResponse{
		Invoice: invoiceToProto(invoice),
		Status:  &statusPB,
	}, nil
}

func invoiceToPB(request *pb.UploadRequest) (*dto.Invoice, error) {
	id, err := uuid.FromBytes(request.Invoice.Id.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice id: %w", err)
	}

	customerId, err := uuid.FromBytes(request.Invoice.CustomerId.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid customer id: %w", err)
	}

	return &dto.Invoice{
		ID:         id,
		CustomerID: customerId,
		Amount:     *request.Invoice.Amount,
		Currency:   *request.Invoice.Currency,
		DueDate:    request.Invoice.DueDate.AsTime(),
		CreatedAt:  request.Invoice.CreatedAt.AsTime(),
		UpdatedAt:  request.Invoice.UpdatedAt.AsTime(),
		Items:      itemsToPB(request.Invoice.Items),
		Notes:      *request.Invoice.Notes,
	}, nil
}

func itemsToPB(items []*types.Item) []dto.Item {
	res := make([]dto.Item, len(items))
	for i, item := range items {
		res[i] = itemToPB(item)
	}
	return res
}

func itemToPB(item *types.Item) dto.Item {
	return dto.Item{
		Description: *item.Description,
		Quantity:    *item.Quantity,
		UnitPrice:   *item.UnitPrice,
		Total:       *item.Total,
	}
}
