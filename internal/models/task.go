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

// TaskSummary is a DTO for displaying tasks on the kanban board
// Contains only the fields needed for the card view plus labels
type TaskSummary struct {
	ID       int
	Title    string
	Labels   []*Label
	ColumnID int
	Position int
}

// TaskDetail is a DTO for the full ticket view
// Contains all task information including description and timestamps
type TaskDetail struct {
	ID          int
	Title       string
	Description string
	Labels      []*Label
	ColumnID    int
	Position    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
