-- name: ScheduleMessage :exec
insert into outbox (payload, topic, next_send_at)
values ($1, $2, $3);

-- name: GetMessages :many
select id, payload, topic from outbox
where next_send_at>=$1
limit $2
for update;

-- name: IncreaseNextSendAt :exec
update outbox
set next_send_at = current_timestamp + (sqlc.arg(time_to_add_sec)::bigint || ' seconds')::interval
where id = ANY(sqlc.arg(ids)::bigint[]);

-- name: DeleteMessage :exec
delete from outbox
where id = $1;