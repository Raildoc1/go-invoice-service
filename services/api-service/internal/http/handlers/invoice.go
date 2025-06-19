package handlers

import (
	"context"
	"go-invoice-service/api-service/internal/dto"
	"go-invoice-service/common/pkg/http/utils"
	"go-invoice-service/common/pkg/logging"
	"go-invoice-service/common/protocol/apiservice/client"
	"go.uber.org/zap"
	"net/http"
)

type InvoiceController interface {
	UploadNew(ctx context.Context, invoice dto.Invoice) error
}

type Invoice struct {
	controller InvoiceController
	logger     *logging.ZapLogger
}

func NewInvoice(controller InvoiceController, logger *logging.ZapLogger) *Invoice {
	return &Invoice{
		controller: controller,
		logger:     logger,
	}
}

func (h *Invoice) UploadNew(w http.ResponseWriter, r *http.Request) {
	requestJSON, err := utils.DecodeJSON[client.UploadInvoiceRequest](r.Body)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to decode request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.controller.UploadNew(r.Context(), convertInvoice(requestJSON.Invoice))
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to upload invoice", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func convertInvoice(invoice client.Invoice) dto.Invoice {
	return dto.Invoice{
		ID:         invoice.ID,
		CustomerID: invoice.CustomerID,
		Amount:     invoice.Amount,
		Currency:   invoice.Currency,
		DueDate:    invoice.DueDate,
		CreatedAt:  invoice.CreatedAt,
		UpdatedAt:  invoice.UpdatedAt,
		Items:      convertItems(invoice.Items),
		Notes:      invoice.Notes,
	}
}

func convertItems(items []client.Item) []dto.Item {
	res := make([]dto.Item, len(items))
	for i, item := range items {
		res[i] = convertItem(item)
	}
	return res
}

func convertItem(item client.Item) dto.Item {
	return dto.Item{
		Description: item.Description,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		Total:       item.Total,
	}
}

func (h *Invoice) Get(w http.ResponseWriter, r *http.Request) {}
