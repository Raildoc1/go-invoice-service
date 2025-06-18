package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"storage-service/internal/data/postgres/generated/queries"
	"storage-service/internal/dto"
)

type Invoice struct {
	qs *queries.Queries
}

func NewInvoice(dbtx queries.DBTX) *Invoice {
	return &Invoice{
		qs: queries.New(dbtx),
	}
}

func (r *Invoice) Add(ctx context.Context, tx *sql.Tx, invoice dto.Invoice, status dto.InvoiceStatus) error {
	qs := r.qs.WithTx(tx)

	err := qs.AddInvoice(ctx, convert(invoice, status))
	if err != nil {
		return fmt.Errorf("add invoice query failed: %w", err)
	}

	return nil
}

func convert(invoice dto.Invoice, status dto.InvoiceStatus) queries.AddInvoiceParams {
	return queries.AddInvoiceParams{
		ID:         invoice.ID,
		CustomerID: invoice.CustomerID,
		Amount:     invoice.Amount,
		Currency:   invoice.Currency,
		DueData:    invoice.DueDate,
		CreatedAt:  invoice.CreatedAt,
		UpdatedAt:  invoice.UpdatedAt,
		Notes:      invoice.Notes,
		Status:     string(status),
	}
}
