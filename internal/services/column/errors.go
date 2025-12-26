package column

import "errors"

// Domain errors for column service
var (
	// Validation errors
	ErrEmptyName        = errors.New("column name cannot be empty")
	ErrNameTooLong      = errors.New("column name cannot exceed 50 characters")
	ErrInvalidColumnID  = errors.New("invalid column ID")
	ErrInvalidProjectID = errors.New("invalid project ID")

	// Business logic errors
	ErrColumnNotFound        = errors.New("column not found")
	ErrColumnHasTasks        = errors.New("cannot delete column with tasks")
	ErrCompletedColumnExists = errors.New("a completed column already exists for this project")
)
