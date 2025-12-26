-- Add holds_completed_tasks column to columns table
-- This boolean indicates if tasks in this column are considered "completed"
-- Only one column per project can have holds_completed_tasks = true (enforced by partial unique index)

-- Step 1: Add the holds_completed_tasks column with default value false
ALTER TABLE columns ADD COLUMN holds_completed_tasks BOOLEAN NOT NULL DEFAULT 0;

-- Step 2: Create a unique partial index to ensure only one column per project can be marked as completed
-- This index only includes rows where holds_completed_tasks = 1
CREATE UNIQUE INDEX idx_columns_completed_per_project
ON columns(project_id) WHERE holds_completed_tasks = 1;

-- Step 3: Set existing "Done" or "Completed" columns as holding completed tasks
-- This provides sensible defaults for existing projects
UPDATE columns SET holds_completed_tasks = 1 WHERE name IN ('Done', 'Completed');
