-- +goose Up
-- Add due_date column to tasks table for task deadline tracking
-- Uses TIMESTAMP for PostgreSQL conventions

-- Add nullable due_date column to tasks table
alter table tasks add column due_date timestamp null;

-- +goose Down
-- Drop column

alter table tasks drop column due_date;
