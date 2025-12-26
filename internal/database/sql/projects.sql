-- ============================================================================
-- PROJECT CRUD OPERATIONS
-- ============================================================================

-- name: CreateProjectRecord :one
INSERT INTO projects (name, description)
VALUES (?, ?)
RETURNING *;

-- name: GetProjectByID :one
SELECT
    id,
    name,
    description,
    created_at,
    updated_at
FROM projects WHERE id = ?;

-- name: GetAllProjects :many
SELECT id, name, description, created_at, updated_at FROM projects ORDER BY id;

-- name: UpdateProject :exec
UPDATE projects SET name = ?,
description = ?,
updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = ?;

-- ============================================================================
-- PROJECT COUNTERS
-- ============================================================================

-- name: InitializeProjectCounter :exec
INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1);

-- name: GetProjectTaskCount :one
SELECT COUNT(*)
FROM tasks t
JOIN columns c ON t.column_id = c.id
WHERE c.project_id = ?;

-- name: DeleteProjectCounter :exec
DELETE FROM project_counters WHERE project_id = ?;

-- ============================================================================
-- PROJECT COLUMN MANAGEMENT
-- ============================================================================

-- name: DeleteTasksByProject :exec
DELETE FROM tasks
WHERE column_id IN (SELECT id FROM columns WHERE project_id = ?);

-- name: DeleteColumnsByProject :exec
DELETE FROM columns
WHERE project_id = ?;
