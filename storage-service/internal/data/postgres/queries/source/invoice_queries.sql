-- name: AddInvoice :execresult
insert into invoices (id, customer_id, amount, currency, due_data, created_at, updated_at, notes, status)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9);