-- +goose Up
-- Add archived column to tasks table for archiving tasks
-- Archived tasks are hidden from the board by default but can be shown via filter
-- PostgreSQL uses native BOOLEAN type

-- Add archived column to tasks table
alter table tasks add column archived BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
-- Drop archived column

alter table tasks drop column archived;
