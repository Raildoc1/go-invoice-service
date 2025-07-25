begin transaction;

create table invoices
(
    id          uuid primary key,
    customer_id uuid                                                              not null,
    amount      bigint                                                            not null,
    currency    varchar(10)                                                       not null,
    due_data    date                                                              not null,
    created_at  timestamp                                                         not null,
    updated_at  timestamp                                                         not null,
    notes       text                                                              not null,
    status      varchar(20) check (status in ('Pending', 'Approved', 'Rejected')) not null default 'Pending'
);

create table invoice_items
(
    id          bigint generated always as identity primary key,
    invoice_id  uuid references invoices (id) not null,
    description text                          not null,
    quantity    int                           not null,
    unit_price  bigint                        not null,
    total       bigint                        not null
);

create table outbox
(
    id           bigint generated always as identity primary key,
    payload      jsonb     not null,
    topic        text      not null,
    next_send_at timestamp not null
);

commit;
