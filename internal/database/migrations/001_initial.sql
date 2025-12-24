-- Paso Database Schema
-- SQLite database for project management with kanban boards

-- ============================================================================
-- LOOKUP TABLES (No dependencies)
-- ============================================================================

-- Task types lookup table
CREATE TABLE IF NOT EXISTS types (
    id INTEGER PRIMARY KEY,
    description TEXT NOT NULL UNIQUE
);

-- Seed default types
INSERT OR IGNORE INTO types (id, description) VALUES
    (1, 'task'),
    (2, 'feature');

-- Priority levels lookup table
CREATE TABLE IF NOT EXISTS priorities (
    id INTEGER PRIMARY KEY,
    description TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL
);

-- Seed default priorities
INSERT OR IGNORE INTO priorities (id, description, color) VALUES
    (1, 'trivial', '#3B82F6'),
    (2, 'low', '#22C55E'),
    (3, 'medium', '#EAB308'),
    (4, 'high', '#F97316'),
    (5, 'critical', '#EF4444');

-- Relationship types lookup table
CREATE TABLE IF NOT EXISTS relation_types (
    id INTEGER PRIMARY KEY,
    p_to_c_label TEXT NOT NULL,
    c_to_p_label TEXT NOT NULL,
    color TEXT NOT NULL,
    is_blocking BOOLEAN NOT NULL DEFAULT 0
);

-- Seed default relation types
INSERT OR IGNORE INTO relation_types (id, p_to_c_label, c_to_p_label, color, is_blocking) VALUES
    (1, 'Parent', 'Child', '#6B7280', 0),
    (2, 'Blocked By', 'Blocker', '#EF4444', 1),
    (3, 'Related To', 'Related To', '#3B82F6', 0);

-- ============================================================================
-- CORE TABLES
-- ============================================================================

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Project counters for auto-incrementing ticket numbers
CREATE TABLE IF NOT EXISTS project_counters (
    project_id INTEGER PRIMARY KEY,
    next_ticket_number INTEGER DEFAULT 1,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Columns table (linked list structure for ordering)
CREATE TABLE IF NOT EXISTS columns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    prev_id INTEGER NULL,
    next_id INTEGER NULL,
    project_id INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Labels table
CREATE TABLE IF NOT EXISTS labels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '#7D56F4',
    project_id INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE(name, project_id)
);

-- ============================================================================
-- TASKS TABLE (Depends on columns, types, priorities)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tasks (
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

-- ============================================================================
-- JOIN TABLES
-- ============================================================================

-- Many-to-many: tasks to labels
CREATE TABLE IF NOT EXISTS task_labels (
    task_id INTEGER NOT NULL,
    label_id INTEGER NOT NULL,
    PRIMARY KEY (task_id, label_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (label_id) REFERENCES labels(id) ON DELETE CASCADE
);

-- Many-to-many: task relationships (parent/child, blockers, etc.)
CREATE TABLE IF NOT EXISTS task_subtasks (
    parent_id INTEGER NOT NULL,
    child_id INTEGER NOT NULL,
    relation_type_id INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (parent_id, child_id),
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (child_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (relation_type_id) REFERENCES relation_types(id)
);

-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================

-- Index for efficient task queries by column
CREATE INDEX IF NOT EXISTS idx_tasks_column ON tasks(column_id, position);

-- Index for efficient project-based column queries
CREATE INDEX IF NOT EXISTS idx_columns_project ON columns(project_id);

-- Index for efficient project-based label queries
CREATE INDEX IF NOT EXISTS idx_labels_project ON labels(project_id);

-- Index for efficient label lookups in task_labels
CREATE INDEX IF NOT EXISTS idx_task_labels_label ON task_labels(label_id);

-- Index for efficient parent task lookups
CREATE INDEX IF NOT EXISTS idx_task_subtasks_parent ON task_subtasks(parent_id);

-- Index for efficient child task lookups
CREATE INDEX IF NOT EXISTS idx_task_subtasks_child ON task_subtasks(child_id);
