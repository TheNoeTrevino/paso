-- +goose Up
-- Add due_date column to tasks table for task deadline tracking
-- Stores optional due date for tasks (null means no due date set)

-- Add nullable due_date column to tasks table
alter table tasks add column due_date DATETIME null;

-- +goose Down
-- Drop due_date column

alter table tasks drop column due_date;
