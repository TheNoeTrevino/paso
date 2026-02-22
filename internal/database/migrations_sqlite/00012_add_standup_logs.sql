-- +goose Up
-- Standup logs table for tracking daily progress per project
create table standup_logs (
    id integer primary key autoincrement,
    project_id integer not null,
    content text not null,
    created_at datetime default current_timestamp,
    foreign key (project_id) references projects(id) on delete cascade
);

-- Index for efficient querying by project
create index idx_standup_logs_project_id on standup_logs(project_id);

-- +goose Down
drop index if exists idx_standup_logs_project_id;
drop table if exists standup_logs;
