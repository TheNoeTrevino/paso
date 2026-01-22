package models

import "time"

// TaskEvent represents an immutable audit event on a task
type TaskEvent struct {
	ID        int
	TaskID    int
	Content   string
	Author    string
	CreatedAt time.Time
}
