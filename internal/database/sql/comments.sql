-- name: CreateComment :one
-- Creates a new comment for a task
insert into task_comments (task_id, content, author)
values (?, ?, ?)
returning *;

-- name: GetComment :one
-- Retrieves a single comment by ID
select
id,
task_id,
content,
author,
created_at,
updated_at
from task_comments
where id = ?;

-- name: GetCommentsByTask :many
-- Retrieves all comments for a task, ordered by creation time (newest first)
select id, task_id, content, author, created_at, updated_at
from task_comments
where task_id = ?
order by created_at desc;

-- name: UpdateComment :exec
-- Updates a comment's content
update task_comments
set content = ?, updated_at = current_timestamp
where id = ?;

-- name: DeleteComment :exec
-- Deletes a comment by ID
delete from task_comments where id = ?;

-- name: GetCommentCountByTask :one
-- Returns the number of comments for a task
select count(*) from task_comments where task_id = ?;
