-- name: CreateColumn :one
-- Creates a new column in a project with optional
-- linked list positioning and task type flags
insert into columns (
    name,
    project_id,
    prev_id,
    next_id,
    holds_ready_tasks,
    holds_completed_tasks,
    holds_in_progress_tasks
)
values ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetColumnByID :one
-- Retrieves a column by its ID with all metadata
select
    id,
    name,
    project_id,
    prev_id,
    next_id,
    holds_ready_tasks,
    holds_completed_tasks,
    holds_in_progress_tasks
from columns
where id = $1;

-- name: GetColumnsByProject :many
-- Retrieves all columns for a specific project
select
    id,
    name,
    project_id,
    prev_id,
    next_id,
    holds_ready_tasks,
    holds_completed_tasks,
    holds_in_progress_tasks
from columns
where project_id = $1;

-- name: GetTailColumnForProject :one
-- Retrieves the last column in a project's linked list (where next_id is NULL)
select id
from columns
where next_id is null
    and project_id = $1
limit 1;

-- name: GetColumnNextID :one
-- Retrieves the next column ID in the linked list
select next_id
from columns
where id = $1;

-- name: UpdateColumnName :exec
-- Updates a column's display name
update columns
set name = $1
where id = $2;

-- name: UpdateColumnNextID :exec
-- Updates the next column pointer in the linked list
update columns
set next_id = $1
where id = $2;

-- name: UpdateColumnPrevID :exec
-- Updates the previous column pointer in the linked list
update columns
set prev_id = $1
where id = $2;

-- name: GetColumnLinkedListInfo :one
-- Retrieves the linked list pointers and project ID for a column
select
    prev_id,
    next_id,
    project_id
from columns
where id = $1;

-- name: DeleteColumn :exec
-- Permanently deletes a column by ID
delete from columns
where id = $1;

-- name: DeleteTasksByColumn :exec
-- Deletes all tasks within a specific column
delete from tasks
where column_id = $1;

-- name: ColumnExists :one
-- Checks if a column exists with the given ID
select count(*)
from columns
where id = $1;

-- name: UpdateColumnHoldsReadyTasks :exec
-- Sets whether a column holds ready tasks (tasks without blockers)
update columns
set holds_ready_tasks = $1
where id = $2;

-- name: GetReadyColumnByProject :one
-- Retrieves the column designated for ready tasks in a project
select
    id,
    name,
    project_id,
    prev_id,
    next_id,
    holds_ready_tasks
from columns
where project_id = $1 and holds_ready_tasks = true
limit 1;

-- name: ClearReadyColumnByProject :exec
-- Clears the ready task flag from all columns in a project
update columns
set holds_ready_tasks = false
where project_id = $1
and holds_ready_tasks = true;

-- name: UpdateColumnHoldsCompletedTasks :exec
-- Sets whether a column holds completed tasks
update columns
set holds_completed_tasks = $1
where id = $2;

-- name: GetCompletedColumnByProject :one
-- Retrieves the column designated for completed tasks in a project
select
    id,
    name,
    project_id,
    prev_id,
    next_id,
    holds_completed_tasks
from columns
where project_id = $1 and holds_completed_tasks = true
limit 1;

-- name: ClearCompletedColumnByProject :exec
-- Clears the completed task flag from all columns in a project
update columns
set holds_completed_tasks = false
where project_id = $1
and holds_completed_tasks = true;

-- name: UpdateColumnHoldsInProgressTasks :exec
-- Sets whether a column holds in-progress tasks
update columns
set holds_in_progress_tasks = $1
where id = $2;

-- name: GetInProgressColumnByProject :one
-- Retrieves the column designated for in-progress tasks in a project
select
    id,
    name,
    project_id,
    prev_id,
    next_id,
    holds_in_progress_tasks
from columns
where project_id = $1 and holds_in_progress_tasks = true
limit 1;

-- name: ClearInProgressColumnByProject :exec
-- Clears the in-progress task flag from all columns in a project
update columns
set holds_in_progress_tasks = false
where project_id = $1
and holds_in_progress_tasks = true;
