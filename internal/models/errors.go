package models

import "errors"

// Task-related errors
var (
	// Validation errors
	ErrEmptyTitle      = errors.New("task title cannot be empty")
	ErrTitleTooLong    = errors.New("task title cannot exceed 255 characters")
	ErrInvalidTaskID   = errors.New("invalid task ID")
	ErrInvalidPriority = errors.New("invalid priority ID")
	ErrInvalidType     = errors.New("invalid type ID")
	ErrInvalidPosition = errors.New("invalid position: must be >= 0")

	// Business logic errors
	ErrTaskNotFound              = errors.New("task not found")
	ErrCircularRelation          = errors.New("circular dependency detected")
	ErrDuplicateRelation         = errors.New("relationship already exists")
	ErrSelfRelation              = errors.New("circular dependency: task cannot have a relationship with itself")
	ErrTaskAlreadyInTargetColumn = errors.New("task is already in target column")

	// Comment validation errors
	ErrEmptyCommentMessage   = errors.New("comment message cannot be empty")
	ErrCommentMessageTooLong = errors.New("comment message cannot exceed 1000 characters")
	ErrInvalidCommentID      = errors.New("invalid comment ID")
	ErrCommentNotFound       = errors.New("comment not found")
)

// Column-related errors
var (
	// Validation errors
	ErrEmptyName        = errors.New("name cannot be empty")
	ErrNameTooLong      = errors.New("name cannot exceed 50 characters")
	ErrInvalidColumnID  = errors.New("invalid column ID")
	ErrInvalidProjectID = errors.New("invalid project ID")

	// Business logic errors
	ErrColumnNotFound         = errors.New("column not found")
	ErrColumnHasTasks         = errors.New("cannot delete column with tasks")
	ErrCompletedColumnExists  = errors.New("a completed column already exists for this project")
	ErrReadyColumnExists      = errors.New("a ready column already exists for this project")
	ErrInProgressColumnExists = errors.New("an in-progress column already exists for this project")
)

// Label-related errors
var (
	// Validation errors
	ErrInvalidColor   = errors.New("invalid color format (must be hex color like #FFFFFF)")
	ErrInvalidLabelID = errors.New("invalid label ID")

	// Business logic errors
	ErrLabelNotFound = errors.New("label not found")
)

// Project-related errors
var (
	// Business logic errors
	ErrProjectNotFound   = errors.New("project not found")
	ErrProjectHasColumns = errors.New("cannot delete project with columns")
	ErrProjectHasTasks   = errors.New("cannot delete project with tasks")
)

// Movement-related errors
var (
	// ErrAlreadyFirstTask indicates that the task is already at the top of the column
	ErrAlreadyFirstTask = errors.New("task is already at the top of the column")

	// ErrAlreadyLastTask indicates that the task is already at the bottom of the column
	ErrAlreadyLastTask = errors.New("task is already at the bottom of the column")

	// ErrAlreadyLastColumn indicates that the task is already in the last column
	ErrAlreadyLastColumn = errors.New("task is already in the last column")

	// ErrAlreadyFirstColumn indicates that the task is already in the first column
	ErrAlreadyFirstColumn = errors.New("task is already in the first column")
)
