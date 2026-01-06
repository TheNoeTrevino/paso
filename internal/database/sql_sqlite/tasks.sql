-- name: CreateTask :one
-- Creates a new task with title, description, position, and ticket number
insert into tasks (
    title,
    description,
    column_id,
    position,
    ticket_number)
values (?, ?, ?, ?, ?)
returning *;

-- name: GetTask :one
-- Retrieves basic task information by ID
select
    id,
    title,
    description,
    column_id,
    position,
    created_at,
    updated_at
from tasks
where id = ?;

-- name: GetTasksByColumn :many
-- Retrieves all tasks in a column, ordered by position
select
    id,
    title,
    description,
    column_id,
    position,
    created_at,
    updated_at
from tasks
where column_id = ?
order by position;

-- name: GetTaskCountByColumn :one
-- Returns the number of tasks in a specific column
select count(*)
from tasks where column_id = ?;

-- name: UpdateTask :exec
-- Updates a task's title and description
update tasks
set title = ?, description = ?, updated_at = current_timestamp
where id = ?;

-- name: UpdateTaskPriority :exec
-- Updates a task's priority level
update tasks
set priority_id = ?, updated_at = current_timestamp
where id = ?;

-- name: UpdateTaskType :exec
-- Updates a task's type classification
update tasks
set type_id = ?, updated_at = current_timestamp
where id = ?;

-- name: DeleteTask :exec
-- Permanently deletes a task by ID
delete from tasks
where id = ?;

-- name: GetTaskDetail :one
-- Retrieves comprehensive task details including:
-- type, priority, column, project, and blocking status
select
    t.id,
    t.title,
    t.description,
    t.column_id,
    t.position,
    t.ticket_number,
    t.created_at,
    t.updated_at,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    c.name as column_name,
    proj.name as project_name,
    exists(
        select 1 from task_subtasks ts
        inner join relation_types rt on ts.relation_type_id = rt.id
        where ts.parent_id = t.id and rt.is_blocking = 1
    ) as is_blocked
from tasks t
inner join columns c on t.column_id = c.id
inner join projects proj on c.project_id = proj.id
left join types ty on t.type_id = ty.id
left join priorities p on t.priority_id = p.id
where t.id = ?;

-- name: GetTaskLabels :many
-- Retrieves all labels attached to a specific task
select l.id, l.name, l.color, l.project_id
from labels l
inner join task_labels tl on l.id = tl.label_id
where tl.task_id = ?
order by l.name;

-- name: GetTaskSummariesByColumn :many
-- Retrieves task summaries with aggregated labels for a specific column using GROUP_CONCAT to avoid N+1 queries
select
    t.id,
    t.title,
    t.column_id,
    t.position,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    cast(coalesce(group_concat(l.id, char(31)), '') as text) as label_ids,
    cast(coalesce(group_concat(l.name, char(31)), '') as text) as label_names,
    cast(coalesce(group_concat(l.color, char(31)), '') as text) as label_colors
from tasks t
left join types ty on t.type_id = ty.id
left join priorities p on t.priority_id = p.id
left join task_labels tl on t.id = tl.task_id
left join labels l on tl.label_id = l.id
where t.column_id = ?
group by
    t.id,
    t.title,
    t.column_id,
    t.position,
    ty.description,
    p.description,
    p.color
order by t.position;

-- name: GetTaskSummariesByProject :many
-- Retrieves task summaries with aggregated labels and blocking status
-- for all tasks in a project
select
    t.id,
    t.title,
    t.column_id,
    t.position,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    cast(coalesce(group_concat(l.id, char(31)), '') as text) as label_ids,
    cast(coalesce(group_concat(l.name, char(31)), '') as text) as label_names,
    cast(coalesce(group_concat(l.color, char(31)), '') as text) as label_colors,
    exists(
        select 1
        from task_subtasks ts
        inner join relation_types rt on ts.relation_type_id = rt.id
        where ts.parent_id = t.id and rt.is_blocking = 1
    ) as is_blocked
from tasks t
inner join columns c on t.column_id = c.id
left join types ty on t.type_id = ty.id
left join priorities p on t.priority_id = p.id
left join task_labels tl on t.id = tl.task_id
left join labels l on tl.label_id = l.id
where c.project_id = ?
group by
    t.id,
    t.title,
    t.column_id,
    t.position,
    ty.description,
    p.description,
    p.color
order by t.position;

-- name: GetReadyTaskSummariesByProject :many
-- Retrieves task summaries for ready tasks (tasks in columns marked as holds_ready_tasks)
select
    t.id,
    t.title,
    t.column_id,
    t.position,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    cast(coalesce(group_concat(l.id, char(31)), '') as text) as label_ids,
    cast(coalesce(group_concat(l.name, char(31)), '') as text) as label_names,
    cast(coalesce(group_concat(l.color, char(31)), '') as text) as label_colors,
    exists(
        select 1
        from task_subtasks ts
        inner join relation_types rt on ts.relation_type_id = rt.id
        where ts.parent_id = t.id and rt.is_blocking = 1
    ) as is_blocked
from tasks t
inner join columns c on t.column_id = c.id
left join types ty on t.type_id = ty.id
left join priorities p on t.priority_id = p.id
left join task_labels tl on t.id = tl.task_id
left join labels l on tl.label_id = l.id
where c.project_id = ? and c.holds_ready_tasks = 1
group by
    t.id,
    t.title,
    t.column_id,
    t.position,
    ty.description,
    p.description,
    p.color
order by t.position;

-- name: GetInProgressTasksByProject :many
-- Retrieves basic information for tasks currently in progress for a project
select
    t.id,
    t.ticket_number,
    t.title,
    t.description,
    c.name as column_name,
    proj.name as project_name
from tasks t
inner join columns c on t.column_id = c.id
inner join projects proj on c.project_id = proj.id
where proj.id = ? and c.holds_in_progress_tasks = 1
order by t.position;

-- name: GetInProgressTaskDetails :many
-- Retrieves comprehensive details for all in-progress tasks using GROUP_CONCAT to avoid N+1 queries
select
    t.id,
    t.ticket_number,
    t.title,
    t.description,
    t.column_id,
    t.position,
    t.created_at,
    t.updated_at,
    c.name as column_name,
    proj.name as project_name,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    cast(coalesce(group_concat(l.id, char(31)), '') as text) as label_ids,
    cast(coalesce(group_concat(l.name, char(31)), '') as text) as label_names,
    cast(coalesce(group_concat(l.color, char(31)), '') as text) as label_colors,
    exists(
        select 1
        from task_subtasks ts
        inner join relation_types rt on ts.relation_type_id = rt.id
        where ts.parent_id = t.id and rt.is_blocking = 1
    ) as is_blocked
from tasks t
inner join columns c on t.column_id = c.id
inner join projects proj on c.project_id = proj.id
left join types ty on t.type_id = ty.id
left join priorities p on t.priority_id = p.id
left join task_labels tl on t.id = tl.task_id
left join labels l on tl.label_id = l.id
where proj.id = ? and c.holds_in_progress_tasks = 1
group by
    t.id,
    t.ticket_number,
    t.title,
    t.description,
    t.column_id,
    t.position,
    t.created_at,
    t.updated_at,
    c.name,
    proj.name,
    ty.description,
    p.description,
    p.color
order by t.position;

-- name: GetTaskSummariesByProjectFiltered :many
-- Retrieves task summaries filtered by title search pattern with aggregated labels
select
    t.id,
    t.title,
    t.column_id,
    t.position,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    cast(coalesce(group_concat(l.id, char(31)), '') as text) as label_ids,
    cast(coalesce(group_concat(l.name, char(31)), '') as text) as label_names,
    cast(coalesce(group_concat(l.color, char(31)), '') as text) as label_colors,
    exists(
        select 1
        from task_subtasks ts
        inner join relation_types rt on ts.relation_type_id = rt.id
        where ts.parent_id = t.id and rt.is_blocking = 1
    ) as is_blocked
from tasks t
inner join columns c on t.column_id = c.id
left join types ty on t.type_id = ty.id
left join priorities p on t.priority_id = p.id
left join task_labels tl on t.id = tl.task_id
left join labels l on tl.label_id = l.id
where c.project_id = ? and t.title like ?
group by
    t.id,
    t.title,
    t.column_id,
    t.position,
    ty.description,
    p.description,
    p.color
order by t.position;

-- name: GetTaskPosition :one
-- Retrieves the current column and position of a task
select column_id, position
from tasks
where id = ?;

-- name: GetNextColumnID :one
-- Retrieves the ID of the next column in the linked list
select next_id
from columns where id = ?;

-- name: GetPrevColumnID :one
-- Retrieves the ID of the previous column in the linked list
select prev_id from columns where id = ?;

-- name: MoveTaskToColumn :exec
-- Moves a task to a different column and updates its position
update tasks
set column_id = ?,
    position = ?,
    updated_at = current_timestamp
where id = ?;

-- name: SetTaskPosition :exec
-- Updates a task's position within its current column
update tasks
set position = ?,
updated_at = current_timestamp
where id = ?;

-- name: SetTaskPositionTemporary :exec
-- Sets task position to -1 temporarily during reordering operations
update tasks 
set position = -1,
updated_at = current_timestamp
where id = ?;

-- name: GetTaskAbove :one
-- Retrieves the task immediately above the given position in a column
select id, position 
from tasks
where column_id = ? and position < ?
order by position desc limit 1;

-- name: GetTaskBelow :one
-- Retrieves the task immediately below the given position in a column
select id, position 
from tasks
where column_id = ? and position > ?
order by position asc limit 1;

-- name: GetProjectIDFromTask :one
-- Retrieves the project ID for a given task by joining through its column
select c.project_id
from tasks t
inner join columns c on t.column_id = c.id
where t.id = ?;

-- name: GetProjectIDFromColumn :one
-- Retrieves the project ID for a given column
select project_id
from columns where id = ?;

-- name: GetNextTicketNumber :one
-- Retrieves the next available ticket number for a project
select next_ticket_number
from project_counters where project_id = ?;

-- name: IncrementTicketNumber :exec
-- Increments the ticket counter for a project after assigning a ticket number
update project_counters
set next_ticket_number = next_ticket_number + 1
where project_id = ?;

-- name: GetParentTasks :many
-- Retrieves all parent tasks for a given child task with relationship details
select t.id, t.ticket_number, t.title, p.name,
rt.id, rt.p_to_c_label, rt.color, rt.is_blocking
from tasks t
inner join task_subtasks ts on t.id = ts.parent_id
inner join relation_types rt on ts.relation_type_id = rt.id
inner join columns c on t.column_id = c.id
inner join projects p on c.project_id = p.id
where ts.child_id = ?
order by p.name, t.ticket_number;

-- name: GetChildTasks :many
-- Retrieves all child tasks for a given parent task with relationship details
select t.id, t.ticket_number, t.title, p.name,
rt.id, rt.c_to_p_label, rt.color, rt.is_blocking
from tasks t
inner join task_subtasks ts on t.id = ts.child_id
inner join relation_types rt on ts.relation_type_id = rt.id
inner join columns c on t.column_id = c.id
inner join projects p on c.project_id = p.id
where ts.parent_id = ?
order by p.name, t.ticket_number;

-- name: GetTaskReferencesForProject :many
-- Retrieves basic task references for all tasks in a project
select t.id, t.ticket_number, t.title, p.name
from tasks t
inner join columns c on t.column_id = c.id
inner join projects p on c.project_id = p.id
where p.id = ?
order by p.name, t.ticket_number;

-- name: AddSubtask :exec
-- Creates a parent-child relationship between two tasks (ignores duplicates)
insert or ignore into
task_subtasks (parent_id, child_id)
values (?, ?);

-- name: AddSubtaskWithRelationType :exec
-- Creates or updates a parent-child relationship with a specific relation type
insert or replace into task_subtasks (parent_id, child_id, relation_type_id)
values (?, ?, ?);

-- name: RemoveSubtask :exec
-- Removes a parent-child relationship between two tasks
delete from task_subtasks where parent_id = ? and child_id = ?;

-- name: GetAllRelationTypes :many
-- Retrieves all available relationship types for task links
select id, p_to_c_label, c_to_p_label, color, is_blocking
from relation_types
order by id;

-- name: GetAllPriorities :many
-- Retrieves all available priority levels
select id, description, color from priorities order by id;

-- name: GetAllTypes :many
-- Retrieves all available task types
select id, description from types order by id;

-- name: GetTasksForTree :many
-- Retrieves all tasks in a project with column
-- and project names for tree visualization
select
    t.id,
    t.ticket_number,
    t.title,
    c.name as column_name,
    proj.name as project_name
from tasks t
inner join columns c on t.column_id = c.id
inner join projects proj on c.project_id = proj.id
where proj.id = ?
order by t.ticket_number;

-- name: GetTaskRelationsForProject :many
-- Retrieves all parent-child task relationships
-- in a project for tree visualization
select
    ts.parent_id,
    ts.child_id,
    rt.c_to_p_label as relation_label,
    rt.color as relation_color,
    rt.is_blocking
from task_subtasks ts
inner join relation_types rt on ts.relation_type_id = rt.id
inner join tasks t_parent on ts.parent_id = t_parent.id
inner join columns c on t_parent.column_id = c.id
where c.project_id = ?;
