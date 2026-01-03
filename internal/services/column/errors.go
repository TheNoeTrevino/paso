package column

import "errors"

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
