-- Add UNIQUE constraint on task position per column
-- This prevents race conditions where multiple tasks can have the same position
-- The constraint is checked after each transaction commits

-- Since SQLite doesn't support ADD CONSTRAINT directly,
-- we need to recreate the table
-- This migration preserves all existing data

-- Step 1: Create new tasks table with UNIQUE constraint
CREATE TABLE IF NOT EXISTS tasks_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    column_id INTEGER NOT NULL,
    position INTEGER NOT NULL,
    ticket_number INTEGER,
    type_id INTEGER NOT NULL DEFAULT 1,
    priority_id INTEGER NOT NULL DEFAULT 3,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE CASCADE,
    FOREIGN KEY (type_id) REFERENCES types(id),
    FOREIGN KEY (priority_id) REFERENCES priorities(id),
    UNIQUE(column_id, position)
);

-- Step 2: Copy data from old table to new table
INSERT INTO tasks_new (
    id,
    title,
    description,
    column_id,
    position,
    ticket_number,
    type_id,
    priority_id,
    created_at,
    updated_at
)
SELECT
    id,
    title,
    description,
    column_id,
    position,
    ticket_number,
    type_id,
    priority_id,
    created_at,
    updated_at
FROM tasks;

-- Step 3: Drop old table
DROP TABLE tasks;

-- Step 4: Rename new table to tasks
ALTER TABLE tasks_new RENAME TO tasks;

-- Step 5: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_tasks_column ON tasks(column_id, position);
