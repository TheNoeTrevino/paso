-- +goose Up
ALTER TABLE tasks RENAME COLUMN ticket_number TO task_number;
ALTER TABLE project_counters RENAME COLUMN next_ticket_number TO next_task_number;

-- +goose Down
ALTER TABLE project_counters RENAME COLUMN next_task_number TO next_ticket_number;
ALTER TABLE tasks RENAME COLUMN task_number TO ticket_number;