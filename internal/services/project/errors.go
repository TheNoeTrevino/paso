package project

import "errors"

// Domain errors for project service
var (
	// Validation errors
	ErrEmptyName        = errors.New("project name cannot be empty")
	ErrNameTooLong      = errors.New("project name cannot exceed 100 characters")
	ErrInvalidProjectID = errors.New("invalid project ID")

	// Business logic errors
	ErrProjectNotFound   = errors.New("project not found")
	ErrProjectHasColumns = errors.New("cannot delete project with columns")
	ErrProjectHasTasks   = errors.New("cannot delete project with tasks")
)
