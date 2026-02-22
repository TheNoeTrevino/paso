-- name: CreateStandupLog :one
-- Creates a new standup log entry for a project
insert into standup_logs (project_id, content)
values (?, ?)
returning *;

-- name: GetStandupLog :one
-- Retrieves a single standup log by ID
select id, project_id, content, created_at
from standup_logs
where id = ?;

-- name: GetStandupLogsByProject :many
-- Retrieves all standup logs for a project, ordered by creation time (newest first)
select id, project_id, content, created_at
from standup_logs
where project_id = ?
order by created_at desc;

-- name: DeleteStandupLog :exec
-- Deletes a standup log by ID
delete from standup_logs where id = ?;
