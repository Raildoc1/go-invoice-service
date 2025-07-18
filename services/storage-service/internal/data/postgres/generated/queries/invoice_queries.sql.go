// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: invoice_queries.sql

package queries

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const addInvoice = `-- name: AddInvoice :exec
insert into invoices (id, customer_id, amount, currency, due_data, created_at, updated_at, notes, status)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

type AddInvoiceParams struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Amount     int64
	Currency   string
	DueData    time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Notes      string
	Status     string
}

func (q *Queries) AddInvoice(ctx context.Context, arg AddInvoiceParams) error {
	_, err := q.db.ExecContext(ctx, addInvoice,
		arg.ID,
		arg.CustomerID,
		arg.Amount,
		arg.Currency,
		arg.DueData,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Notes,
		arg.Status,
	)
	return err
}

const addItem = `-- name: AddItem :exec
insert into invoice_items (invoice_id, description, quantity, unit_price, total)
values ($1, $2, $3, $4, $5)
`

type AddItemParams struct {
	InvoiceID   uuid.UUID
	Description string
	Quantity    int32
	UnitPrice   int64
	Total       int64
}

func (q *Queries) AddItem(ctx context.Context, arg AddItemParams) error {
	_, err := q.db.ExecContext(ctx, addItem,
		arg.InvoiceID,
		arg.Description,
		arg.Quantity,
		arg.UnitPrice,
		arg.Total,
	)
	return err
}

const selectInvoice = `-- name: SelectInvoice :one
select customer_id,
       amount,
       currency,
       due_data,
       created_at,
       updated_at,
       notes,
       status
from invoices
where id = $1
`

type SelectInvoiceRow struct {
	CustomerID uuid.UUID
	Amount     int64
	Currency   string
	DueData    time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Notes      string
	Status     string
}

func (q *Queries) SelectInvoice(ctx context.Context, id uuid.UUID) (SelectInvoiceRow, error) {
	row := q.db.QueryRowContext(ctx, selectInvoice, id)
	var i SelectInvoiceRow
	err := row.Scan(
		&i.CustomerID,
		&i.Amount,
		&i.Currency,
		&i.DueData,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Notes,
		&i.Status,
	)
	return i, err
}

const selectInvoiceItems = `-- name: SelectInvoiceItems :many
select description, quantity, unit_price, total
from invoice_items
where invoice_id = $1
`

type SelectInvoiceItemsRow struct {
	Description string
	Quantity    int32
	UnitPrice   int64
	Total       int64
}

func (q *Queries) SelectInvoiceItems(ctx context.Context, invoiceID uuid.UUID) ([]SelectInvoiceItemsRow, error) {
	rows, err := q.db.QueryContext(ctx, selectInvoiceItems, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SelectInvoiceItemsRow
	for rows.Next() {
		var i SelectInvoiceItemsRow
		if err := rows.Scan(
			&i.Description,
			&i.Quantity,
			&i.UnitPrice,
			&i.Total,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateInvoiceStatus = `-- name: UpdateInvoiceStatus :exec
update invoices
set status = $2
where id = $1
`

type UpdateInvoiceStatusParams struct {
	ID     uuid.UUID
	Status string
}

func (q *Queries) UpdateInvoiceStatus(ctx context.Context, arg UpdateInvoiceStatusParams) error {
	_, err := q.db.ExecContext(ctx, updateInvoiceStatus, arg.ID, arg.Status)
	return err
}
