-- +goose Up
-- Add git_branch column for automatic git branch-based project detection
-- Allows one project per git branch with nullable branch field

-- Add nullable git_branch column to projects table
alter table projects add column git_branch text null;

-- Unique partial index to enforce one project per branch (allows multiple NULLs)
create unique index idx_projects_git_branch_unique
    on projects(git_branch) where git_branch is not null;

-- Standard index for fast branch lookups
create index idx_projects_git_branch on projects(git_branch);

-- +goose Down
-- Drop indexes and column in reverse order

drop index if exists idx_projects_git_branch;
drop index if exists idx_projects_git_branch_unique;
alter table projects drop column git_branch;
