-- name: AddInvoice :exec
insert into invoices (id, customer_id, amount, currency, due_data, created_at, updated_at, notes, status)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: SelectInvoice :one
select customer_id,
       amount,
       currency,
       due_data,
       created_at,
       updated_at,
       notes,
       status
from invoices
where id = $1;

-- name: SelectInvoiceItems :many
select description, quantity, unit_price, total
from invoice_items
where invoice_id = $1;

-- name: UpdateInvoiceStatus :exec
update invoices
set status = $2
where id = $1;