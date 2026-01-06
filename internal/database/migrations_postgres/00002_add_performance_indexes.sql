-- +goose Up
-- Add performance indexes for frequently used database queries
-- This migration adds strategic indexes to optimize query performance without
-- adding unnecessary index overhead. Indexes are created for:
-- 1. Foreign key relationships (standard practice for JOINs)
-- 2. Frequently filtered columns (project_id, column_id, task_id)
-- 3. Partial indexes for special column states (ready, completed, in_progress)
-- 4. Composite indexes for common filter+sort patterns

-- Single column indexes for frequent filtering
create index if not exists idx_tasks_column_id on tasks(column_id);
create index if not exists idx_task_labels_task_id on task_labels(task_id);
create index if not exists idx_labels_project_id on labels(project_id);
create index if not exists idx_columns_project_id on columns(project_id);
create index if not exists idx_task_subtasks_child_id on task_subtasks(
    child_id
);
create index if not exists idx_task_comments_task_id on task_comments(task_id);

-- Composite index for efficient task queries (column_id, position)
-- Already exists in initial schema: idx_tasks_column
-- Verify this is the most common filtering pattern
-- Composite indexes help with both where and order by clauses

-- Partial indexes for column type queries (reduces index size and improves queries)
-- These are unique indexes to enforce only one ready/completed/in_progress column per project
create unique index if not exists idx_columns_ready_unique on columns(
    project_id
)
  where holds_ready_tasks = true;

create unique index if not exists idx_columns_completed_unique on columns(project_id)
  where holds_completed_tasks = true;

create unique index if not exists idx_columns_in_progress_unique on columns(project_id)
  where holds_in_progress_tasks = true;

-- Additional indexes for common query patterns discovered in SQLC queries

-- GetTasksByProject queries join tasks through columns
-- This helps with: where c.project_id = ? queries
-- Already covered by idx_columns_project_id

-- GetTaskLabels and label association queries benefit from:
-- Already covered by idx_task_labels_task_id and idx_labels_project_id

-- GetParentTasks and GetChildTasks queries benefit from:
-- Already covered by idx_task_subtasks_child_id and existing idx_task_subtasks_parent

-- Priority optimization: Index for type lookups (less common but helpful for summary queries)
create index if not exists idx_tasks_type_id on tasks(type_id);
create index if not exists idx_tasks_priority_id on tasks(priority_id);

-- +goose Down
-- Drop all added indexes in reverse order

drop index if exists idx_tasks_priority_id;
drop index if exists idx_tasks_type_id;
drop index if exists idx_columns_in_progress_unique;
drop index if exists idx_columns_completed_unique;
drop index if exists idx_columns_ready_unique;
drop index if exists idx_task_comments_task_id;
drop index if exists idx_task_subtasks_child_id;
drop index if exists idx_columns_project_id;
drop index if exists idx_labels_project_id;
drop index if exists idx_task_labels_task_id;
drop index if exists idx_tasks_column_id;
