-- +goose Up
-- Initial schema for paso task management system

-- Lookup table for task types
create table types (
    id integer primary key,
    description text not null unique
);

insert into types (id, description) values
    (1, 'task'),
    (2, 'feature'),
    (3, 'bug');

-- Lookup table for task priorities
create table priorities (
    id integer primary key,
    description text not null unique,
    color text not null
);

insert into priorities (id, description, color) values
    (1, 'trivial', '#3B82F6'),
    (2, 'low', '#22C55E'),
    (3, 'medium', '#EAB308'),
    (4, 'high', '#F97316'),
    (5, 'critical', '#EF4444');

-- Lookup table for task relationship types
create table relation_types (
    id integer primary key,
    p_to_c_label text not null,
    c_to_p_label text not null,
    color text not null,
    is_blocking boolean not null default 0
);

insert into relation_types (
    id, p_to_c_label, c_to_p_label, color, is_blocking
) values
    (1, 'Parent', 'Child', '#6B7280', 0),
    (2, 'Blocked By', 'Blocker', '#EF4444', 1),
    (3, 'Related To', 'Related To', '#3B82F6', 0);

-- Projects table
create table projects (
    id integer primary key autoincrement,
    name text not null,
    description text default '',
    created_at DATETIME default current_timestamp,
    updated_at DATETIME default current_timestamp
);

-- Project task number counters
create table project_counters (
    project_id integer primary key,
    next_task_number integer default 1,
    foreign key (project_id) references projects(id) on delete cascade
);

-- Columns table (linked list structure for board columns)
create table columns (
    id integer primary key autoincrement,
    name text not null,
    prev_id integer null,
    next_id integer null,
    project_id integer not null,
    holds_ready_tasks boolean not null default 0,
    holds_completed_tasks boolean not null default 0,
    holds_in_progress_tasks boolean not null default 0,
    foreign key (project_id) references projects(id) on delete cascade
);

-- Labels table
create table labels (
    id integer primary key autoincrement,
    name text not null,
    color text not null default '#7D56F4',
    project_id integer not null,
    foreign key (project_id) references projects(id) on delete cascade,
    unique(name, project_id)
);

-- Tasks table
create table tasks (
    id integer primary key autoincrement,
    title text not null,
    description text,
    column_id integer not null,
    position integer not null,
    task_number integer,
    type_id integer not null default 1,
    priority_id integer not null default 3,
    created_at DATETIME default current_timestamp,
    updated_at DATETIME default current_timestamp,
    foreign key (column_id) references columns(id) on delete cascade,
    foreign key (type_id) references types(id),
    foreign key (priority_id) references priorities(id)
);

-- Task-labels many-to-many relationship
create table task_labels (
    task_id integer not null,
    label_id integer not null,
    primary key (task_id, label_id),
    foreign key (task_id) references tasks(id) on delete cascade,
    foreign key (label_id) references labels(id) on delete cascade
);

-- Task relationships (parent-child, blocking, etc.)
create table task_subtasks (
    parent_id integer not null,
    child_id integer not null,
    relation_type_id integer not null default 1,
    primary key (parent_id, child_id),
    foreign key (parent_id) references tasks(id) on delete cascade,
    foreign key (child_id) references tasks(id) on delete cascade,
    foreign key (relation_type_id) references relation_types(id)
);

-- Task comments/notes
create table task_comments (
    id integer primary key autoincrement,
    task_id integer not null,
    content text not null check(length(content) <= 1000),
    author text not null default '',
    created_at DATETIME default current_timestamp,
    updated_at DATETIME default current_timestamp,
    foreign key (task_id) references tasks(id) on delete cascade
);

-- Indexes for performance
create index idx_tasks_column on tasks(column_id, position);
create index idx_columns_project on columns(project_id);
create index idx_labels_project on labels(project_id);
create index idx_task_labels_label on task_labels(label_id);
create index idx_task_subtasks_parent on task_subtasks(parent_id);
create index idx_task_subtasks_child on task_subtasks(child_id);
create index idx_task_comments_task on task_comments(task_id);

-- Unique partial indexes for column constraints
create unique index idx_columns_ready_per_project on columns(project_id) where holds_ready_tasks = 1;
create unique index idx_columns_completed_per_project on columns(project_id) where holds_completed_tasks = 1;
create unique index idx_columns_in_progress_per_project on columns(project_id) where holds_in_progress_tasks = 1;

-- +goose Down
-- Drop all tables and indexes in reverse order

drop index if exists idx_columns_in_progress_per_project;
drop index if exists idx_columns_completed_per_project;
drop index if exists idx_columns_ready_per_project;
drop index if exists idx_task_comments_task;
drop index if exists idx_task_subtasks_child;
drop index if exists idx_task_subtasks_parent;
drop index if exists idx_task_labels_label;
drop index if exists idx_labels_project;
drop index if exists idx_columns_project;
drop index if exists idx_tasks_column;

drop table if exists task_comments;
drop table if exists task_subtasks;
drop table if exists task_labels;
drop table if exists tasks;
drop table if exists labels;
drop table if exists columns;
drop table if exists project_counters;
drop table if exists projects;
drop table if exists relation_types;
drop table if exists priorities;
drop table if exists types;
