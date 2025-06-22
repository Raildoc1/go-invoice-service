package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
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

	err := qs.AddInvoice(ctx, invoiceToDB(invoice, status))
	if err != nil {
		return fmt.Errorf("add invoice query failed: %w", err)
	}

	return nil
}

func (r *Invoice) GetInvoice(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*dto.Invoice, dto.InvoiceStatus, error) {
	qs := r.qs.WithTx(tx)

	invoiceRow, err := qs.SelectInvoice(ctx, id)
	if err != nil {
		return nil, dto.StatusNil, fmt.Errorf("get invoice query failed: %w", err)
	}

	itemRows, err := qs.SelectInvoiceItems(ctx, id)
	if err != nil {
		return nil, dto.StatusNil, fmt.Errorf("get invoice items query failed: %w", err)
	}

	return invoiceFromDB(id, invoiceRow, itemRows), dto.InvoiceStatus(invoiceRow.Status), nil
}

func (r *Invoice) SetStatus(ctx context.Context, tx *sql.Tx, id uuid.UUID, status dto.InvoiceStatus) error {
	qs := r.qs.WithTx(tx)

	err := qs.UpdateInvoiceStatus(ctx, createUpdateStatusParams(id, status))
	if err != nil {
		return fmt.Errorf("update status query failed: %w", err)
	}

	return nil
}

func createUpdateStatusParams(id uuid.UUID, status dto.InvoiceStatus) queries.UpdateInvoiceStatusParams {
	return queries.UpdateInvoiceStatusParams{
		ID:     id,
		Status: string(status),
	}
}

func invoiceToDB(invoice dto.Invoice, status dto.InvoiceStatus) queries.AddInvoiceParams {
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

func invoiceFromDB(id uuid.UUID, invoiceRow queries.SelectInvoiceRow, itemRows []queries.SelectInvoiceItemsRow) *dto.Invoice {
	return &dto.Invoice{
		ID:         id,
		CustomerID: invoiceRow.CustomerID,
		Amount:     invoiceRow.Amount,
		Currency:   invoiceRow.Currency,
		DueDate:    invoiceRow.DueData,
		CreatedAt:  invoiceRow.CreatedAt,
		UpdatedAt:  invoiceRow.UpdatedAt,
		Items:      itemsFromDB(itemRows),
		Notes:      invoiceRow.Notes,
	}
}

func itemsFromDB(itemRows []queries.SelectInvoiceItemsRow) []dto.Item {
	items := make([]dto.Item, len(itemRows))

	for i, itemRow := range itemRows {
		items[i] = itemFromDB(itemRow)
	}

	return items
}

func itemFromDB(row queries.SelectInvoiceItemsRow) dto.Item {
	return dto.Item{
		Description: row.Description,
		Quantity:    row.Quantity,
		UnitPrice:   row.UnitPrice,
		Total:       row.Total,
	}
}
