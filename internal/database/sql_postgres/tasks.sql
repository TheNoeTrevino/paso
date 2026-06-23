-- name: CreateTask :one
-- Creates a new task with title, description, position, task number, and assignee
insert into tasks (
    title,
    description,
    column_id,
    position,
    task_number,
    assignee_id,
    estimate,
    due_date)
values ($1, $2, $3, $4, $5, $6, $7, $8)
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
where id = $1;

-- name: GetTaskTypeAndPriorityIDs :one
-- Retrieves only the type_id and priority_id for a task (lightweight query)
select
    type_id,
    priority_id
from tasks
where id = $1;

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
where column_id = $1
order by position;

-- name: GetTaskCountByColumn :one
-- Returns the number of tasks in a specific column
select count(*)
from tasks where column_id = $1;

-- name: UpdateTask :exec
-- Updates a task's title and description
update tasks
set title = $1, description = $2, updated_at = current_timestamp
where id = $3;

-- name: UpdateTaskPriority :exec
-- Updates a task's priority level
update tasks
set priority_id = $1, updated_at = current_timestamp
where id = $2;

-- name: UpdateTaskType :exec
-- Updates a task's type classification
update tasks
set type_id = $1, updated_at = current_timestamp
where id = $2;

-- name: UpdateTaskAssignee :exec
-- Updates a task's assignee
update tasks
set assignee_id = $1, updated_at = current_timestamp
where id = $2;

-- name: UpdateTaskEstimate :exec
-- Updates a task's estimate
update tasks
set estimate = $1, updated_at = current_timestamp
where id = $2;

-- name: UpdateTaskDueDate :exec
-- Updates a task's due date
update tasks
set due_date = $1, updated_at = current_timestamp
where id = $2;

-- name: UpdateTaskArchived :exec
-- Updates a task's archived status
update tasks
set archived = $1, updated_at = current_timestamp
where id = $2;

-- name: DeleteTask :exec
-- Permanently deletes a task by ID
delete from tasks
where id = $1;

-- name: GetTaskDetail :one
-- Retrieves comprehensive task details including:
-- type, priority, column, project, assignee, and blocking status
select
    t.id,
    t.title,
    t.description,
    t.column_id,
    t.position,
    t.task_number,
    t.created_at,
    t.updated_at,
    t.estimate,
    t.due_date,
    t.archived,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    c.name as column_name,
    proj.name as project_name,
    t.assignee_id,
    a.name as assignee_name,
    exists(
        select 1 from task_subtasks ts
        inner join relation_types rt on ts.relation_type_id = rt.id
        inner join tasks blocker on ts.child_id = blocker.id
        inner join columns bc on blocker.column_id = bc.id
        where ts.parent_id = t.id and rt.is_blocking = true and bc.holds_completed_tasks = false
    ) as is_blocked
from tasks t
inner join columns c on t.column_id = c.id
inner join projects proj on c.project_id = proj.id
left join types ty on t.type_id = ty.id
left join priorities p on t.priority_id = p.id
left join assignees a on t.assignee_id = a.id
where t.id = $1;

-- name: GetTaskLabels :many
-- Retrieves all labels attached to a specific task
select l.id, l.name, l.color, l.project_id
from labels l
inner join task_labels tl on l.id = tl.label_id
where tl.task_id = $1
order by l.name;

-- name: GetTaskSummariesByColumn :many
-- Retrieves task summaries with aggregated labels for a specific column using string_agg to avoid N+1 queries
select
    t.id,
    t.title,
    t.column_id,
    t.position,
    t.estimate,
    t.due_date,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    t.assignee_id,
    a.name as assignee_name,
    cast(coalesce(string_agg(l.id::text, chr(31)), '') as text) as label_ids,
    cast(coalesce(string_agg(l.name, chr(31)), '') as text) as label_names,
    cast(coalesce(string_agg(l.color, chr(31)), '') as text) as label_colors
from tasks t
left join types ty on t.type_id = ty.id
left join priorities p on t.priority_id = p.id
left join assignees a on t.assignee_id = a.id
left join task_labels tl on t.id = tl.task_id
left join labels l on tl.label_id = l.id
where t.column_id = $1
group by
    t.id,
    t.title,
    t.column_id,
    t.position,
    t.estimate,
    t.due_date,
    ty.description,
    p.description,
    p.color,
    t.assignee_id,
    a.name
order by t.position;

-- name: GetTaskSummariesByProject :many
-- Retrieves task summaries with aggregated labels and blocking status
-- for all tasks in a project
select
    t.id,
    t.title,
    t.column_id,
    t.position,
    t.estimate,
    t.due_date,
    t.archived,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    t.assignee_id,
    a.name as assignee_name,
    cast(coalesce(string_agg(l.id::text, chr(31)), '') as text) as label_ids,
    cast(coalesce(string_agg(l.name, chr(31)), '') as text) as label_names,
    cast(coalesce(string_agg(l.color, chr(31)), '') as text) as label_colors,
    exists(
        select 1
        from task_subtasks ts
        inner join relation_types rt on ts.relation_type_id = rt.id
        inner join tasks blocker on ts.child_id = blocker.id
        inner join columns bc on blocker.column_id = bc.id
        where ts.parent_id = t.id and rt.is_blocking = true and bc.holds_completed_tasks = false
    ) as is_blocked
from tasks t
inner join columns c on t.column_id = c.id
left join types ty on t.type_id = ty.id
left join priorities p on t.priority_id = p.id
left join assignees a on t.assignee_id = a.id
left join task_labels tl on t.id = tl.task_id
left join labels l on tl.label_id = l.id
where c.project_id = $1
    and t.archived = false
group by
    t.id,
    t.title,
    t.column_id,
    t.position,
    t.estimate,
    t.due_date,
    ty.description,
    p.description,
    p.color,
    t.assignee_id,
    a.name
order by t.position;

-- name: GetReadyTaskSummariesByProject :many
-- Retrieves task summaries for ready tasks (tasks in columns marked as holds_ready_tasks)
select
    t.id,
    t.title,
    t.column_id,
    t.position,
    t.estimate,
    t.due_date,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    t.assignee_id,
    a.name as assignee_name,
    cast(coalesce(string_agg(l.id::text, chr(31)), '') as text) as label_ids,
    cast(coalesce(string_agg(l.name, chr(31)), '') as text) as label_names,
    cast(coalesce(string_agg(l.color, chr(31)), '') as text) as label_colors,
    exists(
        select 1
        from task_subtasks ts
        inner join relation_types rt on ts.relation_type_id = rt.id
        inner join tasks blocker on ts.child_id = blocker.id
        inner join columns bc on blocker.column_id = bc.id
        where ts.parent_id = t.id and rt.is_blocking = true and bc.holds_completed_tasks = false
    ) as is_blocked
from tasks t
inner join columns c on t.column_id = c.id
left join types ty on t.type_id = ty.id
left join priorities p on t.priority_id = p.id
left join assignees a on t.assignee_id = a.id
left join task_labels tl on t.id = tl.task_id
left join labels l on tl.label_id = l.id
where c.project_id = $1 and c.holds_ready_tasks = true
group by
    t.id,
    t.title,
    t.column_id,
    t.position,
    t.estimate,
    t.due_date,
    ty.description,
    p.description,
    p.color,
    t.assignee_id,
    a.name
order by t.position;

-- name: GetInProgressTasksByProject :many
-- Retrieves basic information for tasks currently in progress for a project
select
    t.id,
    t.task_number,
    t.title,
    t.description,
    c.name as column_name,
    proj.name as project_name
from tasks t
inner join columns c on t.column_id = c.id
inner join projects proj on c.project_id = proj.id
where proj.id = $1 and c.holds_in_progress_tasks = true
order by t.position;

-- name: GetInProgressTaskDetails :many
-- Retrieves comprehensive details for all in-progress tasks using string_agg to avoid N+1 queries
select
    t.id,
    t.task_number,
    t.title,
    t.description,
    t.column_id,
    t.position,
    t.created_at,
    t.updated_at,
    t.estimate,
    t.due_date,
    c.name as column_name,
    proj.name as project_name,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    t.assignee_id,
    a.name as assignee_name,
    cast(coalesce(string_agg(l.id::text, chr(31)), '') as text) as label_ids,
    cast(coalesce(string_agg(l.name, chr(31)), '') as text) as label_names,
    cast(coalesce(string_agg(l.color, chr(31)), '') as text) as label_colors,
    exists(
        select 1
        from task_subtasks ts
        inner join relation_types rt on ts.relation_type_id = rt.id
        inner join tasks blocker on ts.child_id = blocker.id
        inner join columns bc on blocker.column_id = bc.id
        where ts.parent_id = t.id and rt.is_blocking = true and bc.holds_completed_tasks = false
    ) as is_blocked
from tasks t
inner join columns c on t.column_id = c.id
inner join projects proj on c.project_id = proj.id
left join types ty on t.type_id = ty.id
left join priorities p on t.priority_id = p.id
left join assignees a on t.assignee_id = a.id
left join task_labels tl on t.id = tl.task_id
left join labels l on tl.label_id = l.id
where proj.id = $1 and c.holds_in_progress_tasks = true
group by
    t.id,
    t.task_number,
    t.title,
    t.description,
    t.column_id,
    t.position,
    t.created_at,
    t.updated_at,
    t.estimate,
    t.due_date,
    c.name,
    proj.name,
    ty.description,
    p.description,
    p.color,
    t.assignee_id,
    a.name
order by t.position;

-- name: GetTaskSummariesWithFilters :many
-- Retrieves task summaries filtered by multiple optional criteria with aggregated labels
select
    t.id,
    t.title,
    t.column_id,
    t.position,
    t.estimate,
    t.due_date,
    t.archived,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    t.assignee_id,
    a.name as assignee_name,
    cast(coalesce(string_agg(l.id::text, chr(31)), '') as text) as label_ids,
    cast(coalesce(string_agg(l.name, chr(31)), '') as text) as label_names,
    cast(coalesce(string_agg(l.color, chr(31)), '') as text) as label_colors,
    exists(
        select 1
        from task_subtasks ts
        inner join relation_types rt on ts.relation_type_id = rt.id
        inner join tasks blocker on ts.child_id = blocker.id
        inner join columns bc on blocker.column_id = bc.id
        where ts.parent_id = t.id and rt.is_blocking = true and bc.holds_completed_tasks = false
    ) as is_blocked
from tasks t
inner join columns c on t.column_id = c.id
left join types ty on t.type_id = ty.id
left join priorities p on t.priority_id = p.id
left join assignees a on t.assignee_id = a.id
left join task_labels tl on t.id = tl.task_id
left join labels l on tl.label_id = l.id
where c.project_id = sqlc.arg('project_id')
    and (sqlc.narg('title_filter')::text is null or t.title like sqlc.narg('title_filter'))
    and (sqlc.narg('priority_id')::bigint is null or p.id = sqlc.narg('priority_id'))
    and (sqlc.narg('type_id')::bigint is null or ty.id = sqlc.narg('type_id'))
    and (sqlc.narg('assignee_id')::bigint is null or t.assignee_id = sqlc.narg('assignee_id') or (sqlc.narg('assignee_id') = -1 and t.assignee_id is null))
    and (sqlc.arg('label_ids_csv')::text = '' or exists (select 1 from task_labels tl2 where tl2.task_id = t.id and position(',' || cast(tl2.label_id as text) || ',' in sqlc.arg('label_ids_csv')) > 0))
    and (sqlc.arg('show_archived')::boolean = true or t.archived = false)
group by
    t.id,
    t.title,
    t.column_id,
    t.position,
    t.estimate,
    t.due_date,
    ty.description,
    p.description,
    p.color,
    t.assignee_id,
    a.name
order by t.position;

-- name: GetTaskPosition :one
-- Retrieves the current column and position of a task
select column_id, position
from tasks
where id = $1;

-- name: GetNextColumnID :one
-- Retrieves the ID of the next column in the linked list
select next_id
from columns where id = $1;

-- name: GetPrevColumnID :one
-- Retrieves the ID of the previous column in the linked list
select prev_id from columns where id = $1;

-- name: MoveTaskToColumn :exec
-- Moves a task to a different column and updates its position
update tasks
set column_id = $1,
    position = $2,
    updated_at = current_timestamp
where id = $3;

-- name: SetTaskPosition :exec
-- Updates a task's position within its current column
update tasks
set position = $1,
updated_at = current_timestamp
where id = $2;

-- name: SetTaskPositionTemporary :exec
-- Sets task position to -1 temporarily during reordering operations
update tasks
set position = -1,
updated_at = current_timestamp
where id = $1;

-- name: GetTaskAbove :one
-- Retrieves the task immediately above the given position in a column
select id, position
from tasks
where column_id = $1 and position < $2
order by position desc limit 1;

-- name: GetTaskBelow :one
-- Retrieves the task immediately below the given position in a column
select id, position
from tasks
where column_id = $1 and position > $2
order by position asc limit 1;

-- name: GetProjectIDFromTask :one
-- Retrieves the project ID for a given task by joining through its column
select c.project_id
from tasks t
inner join columns c on t.column_id = c.id
where t.id = $1;

-- name: GetProjectIDFromColumn :one
-- Retrieves the project ID for a given column
select project_id
from columns where id = $1;

-- name: GetNextTaskNumber :one
-- Retrieves the next available task number for a project
select next_task_number
from project_counters where project_id = $1;

-- name: IncrementTaskNumber :exec
-- Increments the task counter for a project after assigning a task number
update project_counters
set next_task_number = next_task_number + 1
where project_id = $1;

-- name: GetParentTasks :many
-- Retrieves all parent tasks for a given child task with relationship details
select t.id, t.task_number, t.title, p.name,
rt.id, rt.p_to_c_label, rt.color, rt.is_blocking
from tasks t
inner join task_subtasks ts on t.id = ts.parent_id
inner join relation_types rt on ts.relation_type_id = rt.id
inner join columns c on t.column_id = c.id
inner join projects p on c.project_id = p.id
where ts.child_id = $1
order by p.name, t.task_number;

-- name: GetChildTasks :many
-- Retrieves all child tasks for a given parent task with relationship details
select t.id, t.task_number, t.title, p.name,
rt.id, rt.c_to_p_label, rt.color, rt.is_blocking
from tasks t
inner join task_subtasks ts on t.id = ts.child_id
inner join relation_types rt on ts.relation_type_id = rt.id
inner join columns c on t.column_id = c.id
inner join projects p on c.project_id = p.id
where ts.parent_id = $1
order by p.name, t.task_number;

-- name: GetTaskReferencesForProject :many
-- Retrieves basic task references for all tasks in a project
select t.id, t.task_number, t.title, p.name
from tasks t
inner join columns c on t.column_id = c.id
inner join projects p on c.project_id = p.id
where p.id = $1
order by p.name, t.task_number;

-- name: AddSubtask :exec
-- Creates a parent-child relationship between two tasks (ignores duplicates)
insert into task_subtasks (parent_id, child_id)
values ($1, $2)
on CONFLICT (parent_id, child_id) DO NOTHING;

-- name: AddSubtaskWithRelationType :exec
-- Creates or updates a parent-child relationship with a specific relation type
insert into task_subtasks (parent_id, child_id, relation_type_id)
values ($1, $2, $3)
on CONFLICT (parent_id, child_id) DO update set relation_type_id = $3;

-- name: RemoveSubtask :exec
-- Removes a parent-child relationship between two tasks
delete from task_subtasks where parent_id = $1 and child_id = $2;

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
    t.task_number,
    t.title,
    c.name as column_name,
    proj.name as project_name,
    c.holds_completed_tasks as is_completed
from tasks t
inner join columns c on t.column_id = c.id
inner join projects proj on c.project_id = proj.id
where proj.id = $1
order by t.task_number;

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
where c.project_id = $1;
