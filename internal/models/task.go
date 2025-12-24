package models

import "time"

// Task represents a single task in the kanban board
type Task struct {
	ID          int
	Title       string
	Description string
	TypeID      int
	PriorityID  int
	ColumnID    int
	Position    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TaskReference is a lightweight reference to a related task
// Used for displaying parent/child relationships without loading full task details
type TaskReference struct {
	ID             int
	TicketNumber   int
	Title          string
	ProjectName    string
	RelationTypeID int    // FK to relation_types
	RelationLabel  string // The appropriate label (p_to_c or c_to_p based on context)
	RelationColor  string // Hex color for display
	IsBlocking     bool   // Whether this is a blocking relationship
}

// TaskSummary is a DTO for displaying tasks on the kanban board
// Contains only the fields needed for the card view plus labels
type TaskSummary struct {
	ID                  int
	Title               string
	Labels              []*Label
	TypeDescription     string
	PriorityDescription string
	PriorityColor       string
	ColumnID            int
	Position            int
	IsBlocked           bool // True if any child task has is_blocking=true
}

// TaskDetail is a DTO for the full ticket view
// Contains all task information including description and timestamps
type TaskDetail struct {
	ID                  int
	Title               string
	Description         string
	Labels              []*Label
	ParentTasks         []*TaskReference // Tasks that depend on this task
	ChildTasks          []*TaskReference // Tasks this task depends on
	TypeDescription     string
	PriorityDescription string
	PriorityColor       string
	ColumnID            int
	Position            int
	TicketNumber        int // For display "PROJ-12"
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
