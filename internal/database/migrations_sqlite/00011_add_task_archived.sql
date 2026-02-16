-- +goose Up
-- Add archived column to tasks table for archiving tasks
-- Archived tasks are hidden from the board by default but can be shown via filter
-- SQLite uses INTEGER for boolean values (0 = false, 1 = true)

-- Add nullable archived column to tasks table
alter table tasks add column archived INTEGER NOT NULL DEFAULT 0;

-- +goose Down
-- Drop archived column

alter table tasks drop column archived;
