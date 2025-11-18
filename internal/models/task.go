package models

import "time"

// Task represents a single task in the kanban board
type Task struct {
	ID          int
	Title       string
	Description string
	ColumnID    int
	Position    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
