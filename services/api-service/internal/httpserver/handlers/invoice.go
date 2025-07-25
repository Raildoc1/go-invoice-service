package handlers

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go-invoice-service/api-service/internal/dto"
	"go-invoice-service/common/pkg/http/utils"
	"go-invoice-service/common/pkg/logging"
	"go-invoice-service/common/protocol/apiservice/client"
	"go.uber.org/zap"
	"net/http"
)

type StorageService interface {
	Upload(ctx context.Context, invoice dto.Invoice) error
	Get(ctx context.Context, id uuid.UUID) (dto.Invoice, dto.InvoiceStatus, error)
}

type Invoice struct {
	storageService StorageService
	logger         *logging.ZapLogger
}

func NewInvoice(storageService StorageService, logger *logging.ZapLogger) *Invoice {
	return &Invoice{
		storageService: storageService,
		logger:         logger,
	}
}

func (h *Invoice) Upload(w http.ResponseWriter, r *http.Request) {
	requestJSON, err := utils.DecodeJSON[client.UploadInvoiceRequest](r.Body)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to decode request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.storageService.Upload(r.Context(), invoiceFromProtocol(requestJSON.Invoice))
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to upload invoice", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func invoiceFromProtocol(invoice client.Invoice) dto.Invoice {
	return dto.Invoice{
		ID:         invoice.ID,
		CustomerID: invoice.CustomerID,
		Amount:     currencyAmountFromProtocol(invoice.Amount),
		Currency:   invoice.Currency,
		DueDate:    invoice.DueDate,
		CreatedAt:  invoice.CreatedAt,
		UpdatedAt:  invoice.UpdatedAt,
		Items:      itemsFromProtocol(invoice.Items),
		Notes:      invoice.Notes,
	}
}

func itemsFromProtocol(items []client.Item) []dto.Item {
	res := make([]dto.Item, len(items))
	for i, item := range items {
		res[i] = itemFromProtocol(item)
	}
	return res
}

func itemFromProtocol(item client.Item) dto.Item {
	return dto.Item{
		Description: item.Description,
		Quantity:    item.Quantity,
		UnitPrice:   currencyAmountFromProtocol(item.UnitPrice),
		Total:       currencyAmountFromProtocol(item.Total),
	}
}

func currencyAmountFromProtocol(val decimal.Decimal) int64 {
	return val.Mul(decimal.NewFromInt32(1000)).IntPart()
}
func currencyAmountToProtocol(val int64) decimal.Decimal {
	return decimal.NewFromInt(val).Div(decimal.NewFromInt32(1000))
}

func (h *Invoice) Get(w http.ResponseWriter, r *http.Request) {
	requestJSON, err := utils.DecodeJSON[client.GetInvoiceRequest](r.Body)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to decode request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	invoice, status, err := h.storageService.Get(r.Context(), requestJSON.ID)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to get invoice", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	protocolStatus, err := statusToProtocol(status)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to convert status to protocol", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	resp := client.GetInvoiceResponse{
		Invoice: *invoiceToProtocol(&invoice),
		Status:  protocolStatus,
	}

	err = utils.EncodeJSON(w, resp)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to encode response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func statusToProtocol(status dto.InvoiceStatus) (client.InvoiceStatus, error) {
	switch status {
	case dto.StatusPending:
		return client.StatusPending, nil
	case dto.StatusApproved:
		return client.StatusApproved, nil
	case dto.StatusRejected:
		return client.StatusRejected, nil
	}
	return "", fmt.Errorf("invalid status: %s", status)
}

func invoiceToProtocol(invoice *dto.Invoice) *client.Invoice {
	return &client.Invoice{
		ID:         invoice.ID,
		CustomerID: invoice.CustomerID,
		Amount:     currencyAmountToProtocol(invoice.Amount),
		Currency:   invoice.Currency,
		DueDate:    invoice.DueDate,
		CreatedAt:  invoice.CreatedAt,
		UpdatedAt:  invoice.UpdatedAt,
		Items:      itemsToProtocol(invoice.Items),
		Notes:      invoice.Notes,
	}
}

func itemsToProtocol(items []dto.Item) []client.Item {
	res := make([]client.Item, len(items))

	for i, item := range items {
		res[i] = itemToProtocol(item)
	}

	return res
}

func itemToProtocol(item dto.Item) client.Item {
	return client.Item{
		Description: item.Description,
		Quantity:    item.Quantity,
		UnitPrice:   currencyAmountToProtocol(item.UnitPrice),
		Total:       currencyAmountToProtocol(item.Total),
	}
}
