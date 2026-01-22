-- +goose Up
-- Task events table for immutable audit trail
create table task_events (
    id integer primary key autoincrement,
    task_id integer not null,
    content text not null,
    author text not null default '',
    created_at datetime default current_timestamp,
    foreign key (task_id) references tasks(id) on delete cascade
);

-- Index for efficient querying by task
create index idx_task_events_task_id on task_events(task_id);

-- +goose Down
drop index if exists idx_task_events_task_id;
drop table if exists task_events;
