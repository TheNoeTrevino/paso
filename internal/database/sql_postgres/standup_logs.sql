-- name: CreateStandupLog :one
-- Creates a new standup log entry for a project
insert into standup_logs (project_id, content)
values ($1, $2)
returning *;

-- name: GetStandupLog :one
-- Retrieves a single standup log by ID
select id, project_id, content, created_at
from standup_logs
where id = $1;

-- name: GetStandupLogsByProject :many
-- Retrieves all standup logs for a project, ordered by creation time (newest first)
select id, project_id, content, created_at
from standup_logs
where project_id = $1
order by created_at desc;

-- name: GetStandupLogsByProjectAndDateRange :many
-- Retrieves standup logs for a project within a date range, ordered by creation time (newest first)
select id, project_id, content, created_at
from standup_logs
where project_id = $1
  and created_at >= $2
  and created_at < $3
order by created_at desc;

-- name: DeleteStandupLog :exec
-- Deletes a standup log by ID
delete from standup_logs where id = $1;
