package models

import "errors"

// Domain-specific errors for task movement operations
var (
	// ErrAlreadyFirstTask indicates that the task is already at the top of the column
	ErrAlreadyFirstTask = errors.New("task is already at the top of the column")

	// ErrAlreadyLastTask indicates that the task is already at the bottom of the column
	ErrAlreadyLastTask = errors.New("task is already at the bottom of the column")
)
