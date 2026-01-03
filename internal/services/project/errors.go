package project

import "errors"

// Project-related errors
var (
	// Validation errors
	ErrEmptyName        = errors.New("name cannot be empty")
	ErrNameTooLong      = errors.New("name cannot exceed 50 characters")
	ErrInvalidProjectID = errors.New("invalid project ID")

	// Business logic errors
	ErrProjectNotFound   = errors.New("project not found")
	ErrProjectHasColumns = errors.New("cannot delete project with columns")
	ErrProjectHasTasks   = errors.New("cannot delete project with tasks")
)
