package task

import "errors"

// Domain errors for task service
var (
	// Validation errors
	ErrEmptyTitle      = errors.New("task title cannot be empty")
	ErrTitleTooLong    = errors.New("task title cannot exceed 255 characters")
	ErrInvalidColumnID = errors.New("invalid column ID")
	ErrInvalidPosition = errors.New("invalid position: must be >= 0")
	ErrInvalidTaskID   = errors.New("invalid task ID")
	ErrInvalidPriority = errors.New("invalid priority ID")
	ErrInvalidType     = errors.New("invalid type ID")
	ErrInvalidLabelID  = errors.New("invalid label ID")

	// Business logic errors
	ErrTaskNotFound              = errors.New("task not found")
	ErrCircularRelation          = errors.New("circular relationship detected")
	ErrDuplicateRelation         = errors.New("relationship already exists")
	ErrSelfRelation              = errors.New("task cannot have a relationship with itself")
	ErrTaskAlreadyInTargetColumn = errors.New("task is already in target column")
)
