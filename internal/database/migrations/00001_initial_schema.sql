-- +goose Up
-- Initial schema for paso task management system

-- Lookup table for task types
CREATE TABLE types (
    id INTEGER PRIMARY KEY,
    description TEXT NOT NULL UNIQUE
);

INSERT INTO types (id, description) VALUES
    (1, 'task'),
    (2, 'feature'),
    (3, 'bug');

-- Lookup table for task priorities
CREATE TABLE priorities (
    id INTEGER PRIMARY KEY,
    description TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL
);

INSERT INTO priorities (id, description, color) VALUES
    (1, 'trivial', '#3B82F6'),
    (2, 'low', '#22C55E'),
    (3, 'medium', '#EAB308'),
    (4, 'high', '#F97316'),
    (5, 'critical', '#EF4444');

-- Lookup table for task relationship types
CREATE TABLE relation_types (
    id INTEGER PRIMARY KEY,
    p_to_c_label TEXT NOT NULL,
    c_to_p_label TEXT NOT NULL,
    color TEXT NOT NULL,
    is_blocking BOOLEAN NOT NULL DEFAULT 0
);

INSERT INTO relation_types (id, p_to_c_label, c_to_p_label, color, is_blocking) VALUES
    (1, 'Parent', 'Child', '#6B7280', 0),
    (2, 'Blocked By', 'Blocker', '#EF4444', 1),
    (3, 'Related To', 'Related To', '#3B82F6', 0);

-- Projects table
CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Project ticket number counters
CREATE TABLE project_counters (
    project_id INTEGER PRIMARY KEY,
    next_ticket_number INTEGER DEFAULT 1,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Columns table (linked list structure for board columns)
CREATE TABLE columns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    prev_id INTEGER NULL,
    next_id INTEGER NULL,
    project_id INTEGER NOT NULL,
    holds_ready_tasks BOOLEAN NOT NULL DEFAULT 0,
    holds_completed_tasks BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Labels table
CREATE TABLE labels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '#7D56F4',
    project_id INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE(name, project_id)
);

-- Tasks table
CREATE TABLE tasks (
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
    FOREIGN KEY (priority_id) REFERENCES priorities(id)
);

-- Task-labels many-to-many relationship
CREATE TABLE task_labels (
    task_id INTEGER NOT NULL,
    label_id INTEGER NOT NULL,
    PRIMARY KEY (task_id, label_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (label_id) REFERENCES labels(id) ON DELETE CASCADE
);

-- Task relationships (parent-child, blocking, etc.)
CREATE TABLE task_subtasks (
    parent_id INTEGER NOT NULL,
    child_id INTEGER NOT NULL,
    relation_type_id INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (parent_id, child_id),
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (child_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (relation_type_id) REFERENCES relation_types(id)
);

-- Task comments/notes
CREATE TABLE task_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    content TEXT NOT NULL CHECK(length(content) <= 500),
    author TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_tasks_column ON tasks(column_id, position);
CREATE INDEX idx_columns_project ON columns(project_id);
CREATE INDEX idx_labels_project ON labels(project_id);
CREATE INDEX idx_task_labels_label ON task_labels(label_id);
CREATE INDEX idx_task_subtasks_parent ON task_subtasks(parent_id);
CREATE INDEX idx_task_subtasks_child ON task_subtasks(child_id);
CREATE INDEX idx_task_comments_task ON task_comments(task_id);

-- Unique partial indexes for column constraints
CREATE UNIQUE INDEX idx_columns_ready_per_project ON columns(project_id) WHERE holds_ready_tasks = 1;
CREATE UNIQUE INDEX idx_columns_completed_per_project ON columns(project_id) WHERE holds_completed_tasks = 1;

-- +goose Down
-- Drop all tables and indexes in reverse order

DROP INDEX IF EXISTS idx_columns_completed_per_project;
DROP INDEX IF EXISTS idx_columns_ready_per_project;
DROP INDEX IF EXISTS idx_task_comments_task;
DROP INDEX IF EXISTS idx_task_subtasks_child;
DROP INDEX IF EXISTS idx_task_subtasks_parent;
DROP INDEX IF EXISTS idx_task_labels_label;
DROP INDEX IF EXISTS idx_labels_project;
DROP INDEX IF EXISTS idx_columns_project;
DROP INDEX IF EXISTS idx_tasks_column;

DROP TABLE IF EXISTS task_comments;
DROP TABLE IF EXISTS task_subtasks;
DROP TABLE IF EXISTS task_labels;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS labels;
DROP TABLE IF EXISTS columns;
DROP TABLE IF EXISTS project_counters;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS relation_types;
DROP TABLE IF EXISTS priorities;
DROP TABLE IF EXISTS types;
