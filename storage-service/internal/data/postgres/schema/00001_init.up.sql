begin transaction;

create table invoices
(
    id          uuid primary key,
    customer_id uuid        not null,
    amount      bigint      not null,
    currency    varchar(10) not null,
    due_data    date        not null,
    created_at  timestamp   not null,
    updated_at  timestamp   not null,
    notes       text,
    status      varchar(20) check (status in ('Pending', 'Approved', 'Rejected'))
);

create table invoice_items
(
    id          uuid primary key,
    invoice_id  uuid references invoices (id),
    description text,
    quantity    int,
    unit_price  bigint,
    total       bigint
);

create table new_invoice_outbox
(
    id         int generated always as identity primary key,
    invoice_id uuid references invoices (id),
    retry_time timestamp
);

commit;
