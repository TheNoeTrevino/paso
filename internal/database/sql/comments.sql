-- ============================================================================
-- TASK COMMENT CRUD OPERATIONS
-- ============================================================================

-- name: CreateComment :one
-- Creates a new comment for a task
INSERT INTO task_comments (task_id, content)
VALUES (?, ?)
RETURNING *;

-- name: GetComment :one
-- Gets a single comment by ID
SELECT id, task_id, content, created_at, updated_at
FROM task_comments
WHERE id = ?;

-- name: GetCommentsByTask :many
-- Gets all comments for a task, ordered by creation time (newest first)
SELECT id, task_id, content, created_at, updated_at
FROM task_comments
WHERE task_id = ?
ORDER BY created_at DESC;

-- name: UpdateComment :exec
-- Updates a comment's content
UPDATE task_comments
SET content = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteComment :exec
-- Deletes a comment by ID
DELETE FROM task_comments WHERE id = ?;

-- name: GetCommentCountByTask :one
-- Gets the count of comments for a task
SELECT COUNT(*) FROM task_comments WHERE task_id = ?;
