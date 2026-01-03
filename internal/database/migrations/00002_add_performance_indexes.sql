-- +goose Up
-- Add performance indexes for frequently used database queries
-- This migration adds strategic indexes to optimize query performance without
-- adding unnecessary index overhead. Indexes are created for:
-- 1. Foreign key relationships (standard practice for JOINs)
-- 2. Frequently filtered columns (project_id, column_id, task_id)
-- 3. Partial indexes for special column states (ready, completed, in_progress)
-- 4. Composite indexes for common filter+sort patterns

-- Note: idx_tasks_project_id would require a generated/computed column which SQLite doesn't support
-- in older versions. The column_id index is sufficient for most queries that need project info,
-- as they JOIN through columns anyway.

-- Single column indexes for frequent filtering
CREATE INDEX IF NOT EXISTS idx_tasks_column_id ON tasks(column_id);
CREATE INDEX IF NOT EXISTS idx_task_labels_task_id ON task_labels(task_id);
CREATE INDEX IF NOT EXISTS idx_labels_project_id ON labels(project_id);
CREATE INDEX IF NOT EXISTS idx_columns_project_id ON columns(project_id);
CREATE INDEX IF NOT EXISTS idx_task_subtasks_child_id ON task_subtasks(child_id);
CREATE INDEX IF NOT EXISTS idx_task_comments_task_id ON task_comments(task_id);

-- Composite index for efficient task queries (column_id, position)
-- Already exists in initial schema: idx_tasks_column
-- Verify this is the most common filtering pattern
-- Composite indexes help with both WHERE and ORDER BY clauses

-- Partial indexes for column type queries (reduces index size and improves queries)
-- These are unique indexes to enforce only one ready/completed/in_progress column per project
CREATE UNIQUE INDEX IF NOT EXISTS idx_columns_ready_unique ON columns(project_id)
  WHERE holds_ready_tasks = 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_columns_completed_unique ON columns(project_id)
  WHERE holds_completed_tasks = 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_columns_in_progress_unique ON columns(project_id)
  WHERE holds_in_progress_tasks = 1;

-- Additional indexes for common query patterns discovered in SQLC queries

-- GetTasksByProject queries join tasks through columns
-- This helps with: WHERE c.project_id = ? queries
-- Already covered by idx_columns_project_id

-- GetTaskLabels and label association queries benefit from:
-- Already covered by idx_task_labels_task_id and idx_labels_project_id

-- GetParentTasks and GetChildTasks queries benefit from:
-- Already covered by idx_task_subtasks_child_id and existing idx_task_subtasks_parent

-- Priority optimization: Index for type lookups (less common but helpful for summary queries)
CREATE INDEX IF NOT EXISTS idx_tasks_type_id ON tasks(type_id);
CREATE INDEX IF NOT EXISTS idx_tasks_priority_id ON tasks(priority_id);

-- +goose Down
-- Drop all added indexes in reverse order

DROP INDEX IF EXISTS idx_tasks_priority_id;
DROP INDEX IF EXISTS idx_tasks_type_id;
DROP INDEX IF EXISTS idx_columns_in_progress_unique;
DROP INDEX IF EXISTS idx_columns_completed_unique;
DROP INDEX IF EXISTS idx_columns_ready_unique;
DROP INDEX IF EXISTS idx_task_comments_task_id;
DROP INDEX IF EXISTS idx_task_subtasks_child_id;
DROP INDEX IF EXISTS idx_columns_project_id;
DROP INDEX IF EXISTS idx_labels_project_id;
DROP INDEX IF EXISTS idx_task_labels_task_id;
DROP INDEX IF EXISTS idx_tasks_column_id;
