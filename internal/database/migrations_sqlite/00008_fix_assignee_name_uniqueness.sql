-- +goose Up
-- Fix case-insensitive uniqueness: recreate table with COLLATE NOCASE
-- and remove redundant idx_assignees_name index

CREATE TABLE assignees_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE COLLATE NOCASE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO assignees_new (id, name, created_at, updated_at)
    SELECT id, name, created_at, updated_at FROM assignees;

DROP TABLE assignees;
ALTER TABLE assignees_new RENAME TO assignees;

-- Recreate the foreign key relationship index (the FK itself is maintained by the tasks table definition)
-- The redundant idx_assignees_name is intentionally NOT recreated (task 538)

-- +goose Down
CREATE TABLE assignees_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO assignees_old (id, name, created_at, updated_at)
    SELECT id, name, created_at, updated_at FROM assignees;

DROP TABLE assignees;
ALTER TABLE assignees_old RENAME TO assignees;

CREATE INDEX idx_assignees_name ON assignees(name);
