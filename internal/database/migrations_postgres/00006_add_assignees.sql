-- +goose Up
-- Add assignees table for lightweight user identity tracking
-- Assignees are global to the database, not project-scoped

-- Assignees table stores lightweight user identities
create table assignees (
    id serial primary key,
    name text not null unique,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);

-- Index for fast name lookups (used for @ completions and assignee selection)
create index idx_assignees_name on assignees(name);

-- +goose Down
-- Drop indexes and table in reverse order

drop index if exists idx_assignees_name;
drop table if exists assignees;
