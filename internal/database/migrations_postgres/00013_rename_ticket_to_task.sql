-- +goose Up
-- Rename legacy "ticket" terminology to "task"
-- "ticket" was the original term for tasks; this completes the transition
-- so the schema uses consistent "task" naming.

alter table project_counters rename column next_ticket_number to next_task_number;
alter table tasks rename column ticket_number to task_number;

-- +goose Down
-- Revert "task" naming back to legacy "ticket"

alter table tasks rename column task_number to ticket_number;
alter table project_counters rename column next_task_number to next_ticket_number;
