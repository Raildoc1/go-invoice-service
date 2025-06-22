package servers

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-invoice-service/common/protocol/proto/types"
	pb "go-invoice-service/common/protocol/proto/validation"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"storage-service/internal/dto"
	"time"
)

var _ pb.InvoiceStorageServer = (*ValidationServer)(nil)

type ValidationService interface {
	Get(ctx context.Context, id uuid.UUID) (*dto.Invoice, dto.InvoiceStatus, error)
	SetApproved(ctx context.Context, id uuid.UUID) error
	SetRejected(ctx context.Context, id uuid.UUID) error
}

type ValidationServer struct {
	pb.UnimplementedInvoiceStorageServer
	service ValidationService
}

func NewValidationServer(service ValidationService) *ValidationServer {
	return &ValidationServer{
		service: service,
	}
}

func (s *ValidationServer) Get(ctx context.Context, request *pb.GetInvoiceRequest) (*pb.GetInvoiceResponse, error) {
	id, err := uuidFromProto(request.GetId())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve invoice ID: %w", err)
	}
	invoice, status, err := s.service.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	resp, err := createGetInvoiceResponse(invoice, status)
	if err != nil {
		return nil, fmt.Errorf("failed to create response: %w", err)
	}
	return resp, nil
}

func (s *ValidationServer) SetApproved(ctx context.Context, request *pb.SetApprovedRequest) (*emptypb.Empty, error) {
	id, err := uuidFromProto(request.GetId())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve invoice ID: %w", err)
	}
	err = s.service.SetApproved(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to set invoice approved: %w", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *ValidationServer) SetRejected(ctx context.Context, request *pb.SetRejectedRequest) (*emptypb.Empty, error) {
	id, err := uuidFromProto(request.GetId())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve invoice ID: %w", err)
	}
	err = s.service.SetApproved(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to set invoice rejected: %w", err)
	}
	return &emptypb.Empty{}, nil
}

func createGetInvoiceResponse(invoice *dto.Invoice, status dto.InvoiceStatus) (*pb.GetInvoiceResponse, error) {
	statusPb, err := statusToProto(status)
	if err != nil {
		return nil, fmt.Errorf("failed to convert status to proto: %w", err)
	}
	return &pb.GetInvoiceResponse{
		Invoice: invoiceToProto(invoice),
		Status:  &statusPb,
	}, nil
}

func statusToProto(status dto.InvoiceStatus) (types.InvoiceStatus, error) {
	switch status {
	case dto.StatusPending:
		return types.InvoiceStatus_Pending, nil
	case dto.StatusApproved:
		return types.InvoiceStatus_Approved, nil
	case dto.StatusRejected:
		return types.InvoiceStatus_Rejected, nil
	}
	return 0, fmt.Errorf("unknown status: %s", status)
}

func invoiceToProto(invoice *dto.Invoice) *types.Invoice {
	return &types.Invoice{
		Id:         uuidToProto(invoice.ID),
		CustomerId: uuidToProto(invoice.CustomerID),
		Amount:     &invoice.Amount,
		Currency:   &invoice.Currency,
		DueDate:    timeToProto(invoice.DueDate),
		CreatedAt:  timestamppb.New(invoice.CreatedAt),
		UpdatedAt:  timestamppb.New(invoice.UpdatedAt),
		Items:      itemsToProto(invoice.Items),
		Notes:      &invoice.Notes,
	}
}

func itemsToProto(items []dto.Item) []*types.Item {
	res := make([]*types.Item, len(items))

	for i, item := range items {
		res[i] = itemToProto(item)
	}

	return res
}

func itemToProto(item dto.Item) *types.Item {
	return &types.Item{
		Description: &item.Description,
		Quantity:    &item.Quantity,
		UnitPrice:   &item.UnitPrice,
		Total:       &item.Total,
	}
}

func uuidToProto(id uuid.UUID) *types.UUID {
	return &types.UUID{
		Value: id[:],
	}
}

func timeToProto(date time.Time) *timestamppb.Timestamp {
	return &timestamppb.Timestamp{
		Seconds: date.Unix(),
	}
}

func uuidFromProto(id *types.UUID) (uuid.UUID, error) {
	res, err := uuid.FromBytes(id.Value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid: %w", err)
	}
	return res, nil
}
