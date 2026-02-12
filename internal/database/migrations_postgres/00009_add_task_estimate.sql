-- +goose Up
-- Add estimate column to tasks table for time estimate feature
-- Uses VARCHAR for PostgreSQL conventions

-- Add nullable estimate column to tasks table
alter table tasks add column estimate varchar null;

-- +goose Down
-- Drop column

alter table tasks drop column estimate;
