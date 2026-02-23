package types

// Assignee represents a lightweight user identity for task assignment.
type Assignee struct {
	ID        int64
	Name      string
	CreatedAt NullTime
	UpdatedAt NullTime
}

// Column represents a kanban column in the project board.
type Column struct {
	ID                   int64
	Name                 string
	PrevID               NullInt64 // Nullable - previous column in the linked list
	NextID               NullInt64 // Nullable - next column in the linked list
	ProjectID            int64
	HoldsReadyTasks      bool
	HoldsCompletedTasks  bool
	HoldsInProgressTasks bool
}

// Label represents a label that can be attached to tasks.
type Label struct {
	ID        int64
	Name      string
	Color     string
	ProjectID int64
}

// Priority represents a task priority level.
type Priority struct {
	ID          int64
	Description string
	Color       string
}

// Project represents a project containing tasks and columns.
type Project struct {
	ID          int64
	Name        string
	Description NullString
	GitBranch   NullString
	CreatedAt   NullTime
	UpdatedAt   NullTime
}

// ProjectCounter tracks the next ticket number for a project.
type ProjectCounter struct {
	ProjectID        int64
	NextTicketNumber NullInt64
}

// RelationType represents the type of relationship between tasks.
type RelationType struct {
	ID         int64
	PToCLabel  string // Parent-to-child label
	CToPLabel  string // Child-to-parent label
	Color      string
	IsBlocking bool
}

// Task represents a task in the system.
type Task struct {
	ID           int64
	Title        string
	Description  NullString
	ColumnID     int64
	Position     int64
	TicketNumber NullInt64
	TypeID       int64
	PriorityID   int64
	CreatedAt    NullTime
	UpdatedAt    NullTime
	DueDate      NullTime
	Archived     bool
}

// TaskComment represents a comment on a task.
type TaskComment struct {
	ID        int64
	TaskID    int64
	Content   string
	Author    string
	CreatedAt NullTime
	UpdatedAt NullTime
}

// TaskEvent represents an immutable event/audit entry on a task.
type TaskEvent struct {
	ID        int64
	TaskID    int64
	Content   string
	Author    string
	CreatedAt NullTime
}

// TaskLabel represents the many-to-many relationship between tasks and labels.
type TaskLabel struct {
	TaskID  int64
	LabelID int64
}

// TaskSubtask represents the relationship between parent and child tasks.
type TaskSubtask struct {
	ParentID       int64
	ChildID        int64
	RelationTypeID int64
}

// Type represents a task type.
type Type struct {
	ID          int64
	Description string
}

// GetTaskTypeAndPriorityIDsRow represents the result of fetching type and priority IDs.
type GetTaskTypeAndPriorityIDsRow struct {
	TypeID     int64
	PriorityID int64
}
