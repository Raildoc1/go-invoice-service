package controllers

import (
	"context"
	"go-invoice-service/api-service/internal/dto"
)

type Invoice struct{}

func NewInvoice() *Invoice {
	return &Invoice{}
}

func (c *Invoice) UploadNew(ctx context.Context, invoice dto.Invoice) error {

}
