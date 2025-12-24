-- ============================================================================
-- TASK CRUD OPERATIONS
-- ============================================================================

-- name: CreateTask :one
INSERT INTO tasks (title, description, column_id, position, ticket_number)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetTask :one
SELECT id, title, description, column_id, position, created_at, updated_at
FROM tasks
WHERE id = ?;

-- name: GetTasksByColumn :many
SELECT id, title, description, column_id, position, created_at, updated_at
FROM tasks
WHERE column_id = ?
ORDER BY position;

-- name: GetTaskCountByColumn :one
SELECT COUNT(*) FROM tasks WHERE column_id = ?;

-- name: UpdateTask :exec
UPDATE tasks
SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateTaskPriority :exec
UPDATE tasks
SET priority_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateTaskType :exec
UPDATE tasks
SET type_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = ?;

-- ============================================================================
-- TASK DETAIL QUERIES
-- ============================================================================

-- name: GetTaskDetail :one
SELECT t.id, t.title, t.description, t.column_id, t.position, t.ticket_number, t.created_at, t.updated_at, 
       ty.description as type_description, 
       p.description as priority_description, 
       p.color as priority_color
FROM tasks t
LEFT JOIN types ty ON t.type_id = ty.id
LEFT JOIN priorities p ON t.priority_id = p.id
WHERE t.id = ?;

-- name: GetTaskLabels :many
SELECT l.id, l.name, l.color, l.project_id
FROM labels l
INNER JOIN task_labels tl ON l.id = tl.label_id
WHERE tl.task_id = ?
ORDER BY l.name;

-- ============================================================================
-- TASK SUMMARIES (Optimized with JOINs to avoid N+1 queries)
-- ============================================================================

-- name: GetTaskSummariesByColumn :many
SELECT
    t.id,
    t.title,
    t.column_id,
    t.position,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    CAST(COALESCE(GROUP_CONCAT(l.id, CHAR(31)), '') AS TEXT) as label_ids,
    CAST(COALESCE(GROUP_CONCAT(l.name, CHAR(31)), '') AS TEXT) as label_names,
    CAST(COALESCE(GROUP_CONCAT(l.color, CHAR(31)), '') AS TEXT) as label_colors
FROM tasks t
LEFT JOIN types ty ON t.type_id = ty.id
LEFT JOIN priorities p ON t.priority_id = p.id
LEFT JOIN task_labels tl ON t.id = tl.task_id
LEFT JOIN labels l ON tl.label_id = l.id
WHERE t.column_id = ?
GROUP BY t.id, t.title, t.column_id, t.position, ty.description, p.description, p.color
ORDER BY t.position;

-- name: GetTaskSummariesByProject :many
SELECT
    t.id,
    t.title,
    t.column_id,
    t.position,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    CAST(COALESCE(GROUP_CONCAT(l.id, CHAR(31)), '') AS TEXT) as label_ids,
    CAST(COALESCE(GROUP_CONCAT(l.name, CHAR(31)), '') AS TEXT) as label_names,
    CAST(COALESCE(GROUP_CONCAT(l.color, CHAR(31)), '') AS TEXT) as label_colors,
    EXISTS(
        SELECT 1 FROM task_subtasks ts
        INNER JOIN relation_types rt ON ts.relation_type_id = rt.id
        WHERE ts.parent_id = t.id AND rt.is_blocking = 1
    ) as is_blocked
FROM tasks t
INNER JOIN columns c ON t.column_id = c.id
LEFT JOIN types ty ON t.type_id = ty.id
LEFT JOIN priorities p ON t.priority_id = p.id
LEFT JOIN task_labels tl ON t.id = tl.task_id
LEFT JOIN labels l ON tl.label_id = l.id
WHERE c.project_id = ?
GROUP BY t.id, t.title, t.column_id, t.position, ty.description, p.description, p.color
ORDER BY t.position;

-- name: GetTaskSummariesByProjectFiltered :many
SELECT
    t.id,
    t.title,
    t.column_id,
    t.position,
    ty.description as type_description,
    p.description as priority_description,
    p.color as priority_color,
    CAST(COALESCE(GROUP_CONCAT(l.id, CHAR(31)), '') AS TEXT) as label_ids,
    CAST(COALESCE(GROUP_CONCAT(l.name, CHAR(31)), '') AS TEXT) as label_names,
    CAST(COALESCE(GROUP_CONCAT(l.color, CHAR(31)), '') AS TEXT) as label_colors,
    EXISTS(
        SELECT 1 FROM task_subtasks ts
        INNER JOIN relation_types rt ON ts.relation_type_id = rt.id
        WHERE ts.parent_id = t.id AND rt.is_blocking = 1
    ) as is_blocked
FROM tasks t
INNER JOIN columns c ON t.column_id = c.id
LEFT JOIN types ty ON t.type_id = ty.id
LEFT JOIN priorities p ON t.priority_id = p.id
LEFT JOIN task_labels tl ON t.id = tl.task_id
LEFT JOIN labels l ON tl.label_id = l.id
WHERE c.project_id = ? AND t.title LIKE ?
GROUP BY t.id, t.title, t.column_id, t.position, ty.description, p.description, p.color
ORDER BY t.position;

-- ============================================================================
-- TASK MOVEMENT OPERATIONS
-- ============================================================================

-- name: GetTaskPosition :one
SELECT column_id, position FROM tasks WHERE id = ?;

-- name: GetNextColumnID :one
SELECT next_id FROM columns WHERE id = ?;

-- name: GetPrevColumnID :one
SELECT prev_id FROM columns WHERE id = ?;

-- name: MoveTaskToColumn :exec
UPDATE tasks
SET column_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: SetTaskPosition :exec
UPDATE tasks
SET position = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: SetTaskPositionTemporary :exec
UPDATE tasks SET position = -1, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: GetTaskAbove :one
SELECT id, position FROM tasks
WHERE column_id = ? AND position < ?
ORDER BY position DESC LIMIT 1;

-- name: GetTaskBelow :one
SELECT id, position FROM tasks
WHERE column_id = ? AND position > ?
ORDER BY position ASC LIMIT 1;

-- ============================================================================
-- PROJECT HELPERS (for event notifications)
-- ============================================================================

-- name: GetProjectIDFromTask :one
SELECT c.project_id FROM tasks t
INNER JOIN columns c ON t.column_id = c.id
WHERE t.id = ?;

-- name: GetProjectIDFromColumn :one
SELECT project_id FROM columns WHERE id = ?;

-- ============================================================================
-- TICKET NUMBER MANAGEMENT
-- ============================================================================

-- name: GetNextTicketNumber :one
SELECT next_ticket_number FROM project_counters WHERE project_id = ?;

-- name: IncrementTicketNumber :exec
UPDATE project_counters SET next_ticket_number = next_ticket_number + 1 WHERE project_id = ?;

-- ============================================================================
-- TASK RELATIONSHIPS (Subtasks, Parents, Blockers)
-- ============================================================================

-- name: GetParentTasks :many
SELECT t.id, t.ticket_number, t.title, p.name,
       rt.id, rt.p_to_c_label, rt.color, rt.is_blocking
FROM tasks t
INNER JOIN task_subtasks ts ON t.id = ts.parent_id
INNER JOIN relation_types rt ON ts.relation_type_id = rt.id
INNER JOIN columns c ON t.column_id = c.id
INNER JOIN projects p ON c.project_id = p.id
WHERE ts.child_id = ?
ORDER BY p.name, t.ticket_number;

-- name: GetChildTasks :many
SELECT t.id, t.ticket_number, t.title, p.name,
       rt.id, rt.c_to_p_label, rt.color, rt.is_blocking
FROM tasks t
INNER JOIN task_subtasks ts ON t.id = ts.child_id
INNER JOIN relation_types rt ON ts.relation_type_id = rt.id
INNER JOIN columns c ON t.column_id = c.id
INNER JOIN projects p ON c.project_id = p.id
WHERE ts.parent_id = ?
ORDER BY p.name, t.ticket_number;

-- name: GetTaskReferencesForProject :many
SELECT t.id, t.ticket_number, t.title, p.name
FROM tasks t
INNER JOIN columns c ON t.column_id = c.id
INNER JOIN projects p ON c.project_id = p.id
WHERE p.id = ?
ORDER BY p.name, t.ticket_number;

-- name: AddSubtask :exec
INSERT OR IGNORE INTO task_subtasks (parent_id, child_id) VALUES (?, ?);

-- name: AddSubtaskWithRelationType :exec
INSERT OR REPLACE INTO task_subtasks (parent_id, child_id, relation_type_id)
VALUES (?, ?, ?);

-- name: RemoveSubtask :exec
DELETE FROM task_subtasks WHERE parent_id = ? AND child_id = ?;

-- name: GetAllRelationTypes :many
SELECT id, p_to_c_label, c_to_p_label, color, is_blocking
FROM relation_types
ORDER BY id;

-- ============================================================================
-- PRIORITIES AND TYPES
-- ============================================================================

-- name: GetAllPriorities :many
SELECT id, description, color FROM priorities ORDER BY id;

-- name: GetAllTypes :many
SELECT id, description FROM types ORDER BY id;
