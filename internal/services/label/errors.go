package label

import "errors"

// Domain errors for label service
var (
	// Validation errors
	ErrEmptyName        = errors.New("label name cannot be empty")
	ErrNameTooLong      = errors.New("label name cannot exceed 50 characters")
	ErrInvalidColor     = errors.New("invalid color format (must be hex color like #FFFFFF)")
	ErrInvalidLabelID   = errors.New("invalid label ID")
	ErrInvalidProjectID = errors.New("invalid project ID")
	ErrInvalidTaskID    = errors.New("invalid task ID")

	// Business logic errors
	ErrLabelNotFound = errors.New("label not found")
)
