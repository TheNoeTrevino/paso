-- name: CreateLabel :one
-- Creates a new label with name, color, and project association
insert into labels (name, color, project_id)
values ($1, $2, $3)
returning *;

-- name: GetLabelsByProject :many
-- Retrieves all labels for a project, ordered alphabetically by name
select
    id,
    name,
    color,
    project_id
from labels
where project_id = $1
order by name;

-- name: GetLabelByID :one
-- Retrieves a label by its ID
select id, name, color, project_id
from labels
where id = $1;

-- name: GetLabelsForTask :many
-- Retrieves all labels attached to a specific task
select l.id, l.name, l.color, l.project_id
from labels l
inner join task_labels tl on l.id = tl.label_id
where tl.task_id = $1
order by l.name;

-- name: UpdateLabel :exec
-- Updates a label's name and color
update labels set name = $1, color = $2 where id = $3;

-- name: DeleteLabel :exec
-- Permanently deletes a label by ID
delete from labels where id = $1;

-- name: AddLabelToTask :exec
-- Attaches a label to a task (ignores if already attached)
insert into task_labels (task_id, label_id) values ($1, $2)
on CONFLICT (task_id, label_id) DO NOTHING;

-- name: RemoveLabelFromTask :exec
-- Removes a specific label from a task
delete from task_labels where task_id = $1 and label_id = $2;

-- name: DeleteAllLabelsFromTask :exec
-- Removes all labels from a task
delete from task_labels where task_id = $1;

-- name: InsertTaskLabel :exec
-- Creates a task-label association
insert into task_labels (task_id, label_id) values ($1, $2);

-- name: GetLabelCountByProject :one
-- Returns the count of labels for a project
select count(*) from labels where project_id = $1;

-- name: UpsertLabel :exec
-- Inserts a label or ignores if it already exists (for seeding)
insert into labels (name, color, project_id) values ($1, $2, $3)
on conflict (name, project_id) do nothing;

-- name: CountTasksByLabel :one
-- Returns the count of tasks that have this label attached
select count(*) from task_labels where label_id = $1;
