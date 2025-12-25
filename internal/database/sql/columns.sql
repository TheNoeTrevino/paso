-- ============================================================================
-- COLUMN CRUD OPERATIONS
-- ============================================================================

-- name: CreateColumn :one
INSERT INTO columns (name, project_id, prev_id, next_id)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetColumnByID :one
SELECT id, name, project_id, prev_id, next_id
FROM columns
WHERE id = ?;

-- name: GetColumnsByProject :many
SELECT id, name, project_id, prev_id, next_id
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
