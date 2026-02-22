-- name: CreateAssignee :one
-- Creates a new assignee
insert into assignees (name)
values ($1)
returning *;

-- name: GetAssigneeByID :one
-- Retrieves an assignee by ID
select id, name, created_at, updated_at
from assignees
where id = $1;

-- name: GetAssigneeByName :one
-- Retrieves an assignee by name (case-insensitive)
select id, name, created_at, updated_at
from assignees
where lower(name) = lower($1);

-- name: ListAssignees :many
-- Lists all assignees ordered by name
select id, name, created_at, updated_at
from assignees
order by name;

-- name: DeleteAssignee :execrows
-- Deletes an assignee by ID
-- Note: Tasks will have assignee_id set to null via on delete set null
delete from assignees
where id = $1;
