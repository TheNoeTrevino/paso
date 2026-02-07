-- +goose Up
-- Add assignee_id column to tasks table for task assignment feature
-- Uses on delete set null to preserve tasks when assignee is deleted

-- Add nullable assignee_id column to tasks table
alter table tasks add column assignee_id integer null references assignees(id) on delete set null;

-- Index for fast assignee lookups
create index idx_tasks_assignee_id on tasks(assignee_id);

-- +goose Down
-- Drop index and column in reverse order

drop index if exists idx_tasks_assignee_id;
alter table tasks drop column assignee_id;
