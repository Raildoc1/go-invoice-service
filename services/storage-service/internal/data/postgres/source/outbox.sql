-- name: ScheduleMessage :exec
insert into outbox (payload, topic, next_send_at)
values ($1, $2, $3);