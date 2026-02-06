-- +goose Up
-- Fix case-insensitive uniqueness for Postgres
-- Drop redundant index and add functional unique index on lower(name)

DROP INDEX IF EXISTS idx_assignees_name;
ALTER TABLE assignees DROP CONSTRAINT IF EXISTS assignees_name_key;
CREATE UNIQUE INDEX idx_assignees_name_lower ON assignees(lower(name));

-- +goose Down
DROP INDEX IF EXISTS idx_assignees_name_lower;
ALTER TABLE assignees ADD CONSTRAINT assignees_name_key UNIQUE (name);
CREATE INDEX idx_assignees_name ON assignees(name);
