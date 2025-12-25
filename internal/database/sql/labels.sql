-- ============================================================================
-- LABEL CRUD OPERATIONS
-- ============================================================================

-- name: CreateLabel :one
INSERT INTO labels (name, color, project_id)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetLabelsByProject :many
SELECT id, name, color, project_id
FROM labels
WHERE project_id = ?
ORDER BY name;

-- name: GetLabelByID :one
SELECT id, name, color, project_id
FROM labels
WHERE id = ?;

-- name: GetLabelsForTask :many
SELECT l.id, l.name, l.color, l.project_id
FROM labels l
INNER JOIN task_labels tl ON l.id = tl.label_id
WHERE tl.task_id = ?
ORDER BY l.name;

-- name: UpdateLabel :exec
UPDATE labels SET name = ?, color = ? WHERE id = ?;

-- name: DeleteLabel :exec
DELETE FROM labels WHERE id = ?;

-- ============================================================================
-- TASK-LABEL ASSOCIATIONS
-- ============================================================================

-- name: AddLabelToTask :exec
INSERT OR IGNORE INTO task_labels (task_id, label_id) VALUES (?, ?);

-- name: RemoveLabelFromTask :exec
DELETE FROM task_labels WHERE task_id = ? AND label_id = ?;

-- name: DeleteAllLabelsFromTask :exec
DELETE FROM task_labels WHERE task_id = ?;

-- name: InsertTaskLabel :exec
INSERT INTO task_labels (task_id, label_id) VALUES (?, ?);
