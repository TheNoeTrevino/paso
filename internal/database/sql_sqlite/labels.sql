-- name: CreateLabel :one
-- Creates a new label with name, color, and project association
insert into labels (name, color, project_id)
values (?, ?, ?)
returning *;

-- name: GetLabelsByProject :many
-- Retrieves all labels for a project, ordered alphabetically by name
select
    id,
    name,
    color,
    project_id
from labels
where project_id = ?
order by name;

-- name: GetLabelByID :one
-- Retrieves a label by its ID
select id, name, color, project_id
from labels
where id = ?;

-- name: GetLabelsForTask :many
-- Retrieves all labels attached to a specific task
select l.id, l.name, l.color, l.project_id
from labels l
inner join task_labels tl on l.id = tl.label_id
where tl.task_id = ?
order by l.name;

-- name: UpdateLabel :exec
-- Updates a label's name and color
update labels set name = ?, color = ? where id = ?;

-- name: DeleteLabel :exec
-- Permanently deletes a label by ID
delete from labels where id = ?;

-- name: AddLabelToTask :exec
-- Attaches a label to a task (ignores if already attached)
insert or ignore into task_labels (task_id, label_id) values (?, ?);

-- name: RemoveLabelFromTask :exec
-- Removes a specific label from a task
delete from task_labels where task_id = ? and label_id = ?;

-- name: DeleteAllLabelsFromTask :exec
-- Removes all labels from a task
delete from task_labels where task_id = ?;

-- name: InsertTaskLabel :exec
-- Creates a task-label association
insert into task_labels (task_id, label_id) values (?, ?);
