-- +goose Up
-- Add estimate column to tasks table for time estimation feature
-- Stores time estimate in a flexible text format (e.g., "2h", "1d", "30m")

-- Add nullable estimate column to tasks table
alter table tasks add column estimate text null;

-- +goose Down
-- Drop estimate column

alter table tasks drop column estimate;
