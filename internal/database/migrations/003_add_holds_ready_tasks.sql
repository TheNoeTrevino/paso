-- Add holds_ready_tasks column to columns table
-- This boolean indicates if tasks in this column are considered "ready" for work
-- Only one column per project can have holds_ready_tasks = true (enforced by partial unique index)

-- Step 1: Add the holds_ready_tasks column with default value false
ALTER TABLE columns ADD COLUMN holds_ready_tasks BOOLEAN NOT NULL DEFAULT 0;

-- Step 2: Create a unique partial index to ensure only one column per project can be marked as ready
-- This index only includes rows where holds_ready_tasks = 1
CREATE UNIQUE INDEX idx_columns_ready_per_project
ON columns(project_id) WHERE holds_ready_tasks = 1;

-- Step 3: Set existing "Todo" columns as holding ready tasks
-- This provides sensible defaults for existing projects
UPDATE columns SET holds_ready_tasks = 1 WHERE name = 'Todo';
