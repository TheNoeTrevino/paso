-- name: CreateTaskEvent :one
-- Creates a new event for a task
insert into task_events (task_id, content, author)
values ($1, $2, $3)
returning *;

-- name: GetEventsByTask :many
-- Retrieves all events for a task, ordered by creation time (newest first)
select id, task_id, content, author, created_at
from task_events
where task_id = $1
order by created_at desc;

-- name: DeleteEventsByTask :exec
-- Deletes all events for a task (for testing/cleanup)
delete from task_events where task_id = $1;
