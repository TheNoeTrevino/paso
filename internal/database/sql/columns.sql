-- ============================================================================
-- COLUMN CRUD OPERATIONS
-- ============================================================================

-- name: CreateColumn :one
INSERT INTO columns (name, project_id, prev_id, next_id, holds_ready_tasks)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetColumnByID :one
SELECT id, name, project_id, prev_id, next_id, holds_ready_tasks
FROM columns
WHERE id = ?;

-- name: GetColumnsByProject :many
SELECT id, name, project_id, prev_id, next_id, holds_ready_tasks
FROM columns
WHERE project_id = ?;

-- name: GetTailColumnForProject :one
SELECT id FROM columns
WHERE next_id IS NULL AND project_id = ?
LIMIT 1;

-- name: GetColumnNextID :one
SELECT next_id FROM columns WHERE id = ?;

-- name: UpdateColumnName :exec
UPDATE columns SET name = ? WHERE id = ?;

-- name: UpdateColumnNextID :exec
UPDATE columns SET next_id = ? WHERE id = ?;

-- name: UpdateColumnPrevID :exec
UPDATE columns SET prev_id = ? WHERE id = ?;

-- name: GetColumnLinkedListInfo :one
SELECT prev_id, next_id, project_id FROM columns WHERE id = ?;

-- name: DeleteColumn :exec
DELETE FROM columns WHERE id = ?;

-- name: DeleteTasksByColumn :exec
DELETE FROM tasks WHERE column_id = ?;

-- ============================================================================
-- COLUMN VERIFICATION
-- ============================================================================

-- name: ColumnExists :one
SELECT COUNT(*) FROM columns WHERE id = ?;

-- ============================================================================
-- READY COLUMN OPERATIONS
-- ============================================================================

-- name: UpdateColumnHoldsReadyTasks :exec
UPDATE columns SET holds_ready_tasks = ? WHERE id = ?;

-- name: GetReadyColumnByProject :one
SELECT id, name, project_id, prev_id, next_id, holds_ready_tasks
FROM columns
WHERE project_id = ? AND holds_ready_tasks = 1
LIMIT 1;

-- name: ClearReadyColumnByProject :exec
UPDATE columns SET holds_ready_tasks = 0
WHERE project_id = ? AND holds_ready_tasks = 1;
