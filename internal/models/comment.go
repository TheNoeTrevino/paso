package models

import "time"

// Comment represents a note/comment on a task
type Comment struct {
	ID        int
	TaskID    int
	Message   string
	Author    string
	CreatedAt time.Time
}
