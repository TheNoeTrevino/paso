-- Migration: Add task_comments table for storing notes on tasks
-- Each comment has a 500 character limit and is associated with a task

CREATE TABLE IF NOT EXISTS task_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    content TEXT NOT NULL CHECK(length(content) <= 500),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Index for efficient comment retrieval by task
CREATE INDEX IF NOT EXISTS idx_task_comments_task ON task_comments(task_id);
