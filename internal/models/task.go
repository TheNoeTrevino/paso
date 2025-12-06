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

// TaskReference is a lightweight reference to a related task
// Used for displaying parent/child relationships without loading full task details
type TaskReference struct {
	ID           int
	TicketNumber int
	Title        string
	ProjectName  string
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
	ID           int
	Title        string
	Description  string
	Labels       []*Label
	ParentTasks  []*TaskReference // Tasks that depend on this task
	ChildTasks   []*TaskReference // Tasks this task depends on
	ColumnID     int
	Position     int
	TicketNumber int // For display "PROJ-12"
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
