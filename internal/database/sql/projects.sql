-- name: CreateProjectRecord :one
-- Creates a new project with name and description
insert into projects (name, description)
values (?, ?)
returning *;

-- name: GetProjectByID :one
-- Retrieves a project by its ID with all metadata
select
    id,
    name,
    description,
    created_at,
    updated_at
from projects where id = ?;

-- name: GetAllProjects :many
-- Retrieves all projects ordered by ID
select id, name, description, created_at, updated_at from projects order by id;

-- name: UpdateProject :exec
-- Updates a project's name and description
update projects set name = ?,
description = ?,
updated_at = current_timestamp where id = ?;

-- name: DeleteProject :exec
-- Permanently deletes a project by ID
delete from projects where id = ?;

-- name: InitializeProjectCounter :exec
-- Initializes the ticket number counter for a new project starting at 1
insert into project_counters (project_id, next_ticket_number) values (?, 1);

-- name: GetProjectTaskCount :one
-- Returns the total number of tasks in a project
select count(*)
from tasks t
join columns c on t.column_id = c.id
where c.project_id = ?;

-- name: DeleteProjectCounter :exec
-- Deletes the ticket counter for a project
delete from project_counters where project_id = ?;

-- name: DeleteTasksByProject :exec
-- Deletes all tasks belonging to a project
delete from tasks
where column_id in (select id from columns where project_id = ?);

-- name: DeleteColumnsByProject :exec
-- Deletes all columns belonging to a project
delete from columns
where project_id = ?;
