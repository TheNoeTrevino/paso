package models

import "errors"

// Domain-specific errors for task movement operations
var (
	// ErrNoNextColumn indicates that the task is already in the last column
	ErrNoNextColumn = errors.New("task is already in the last column")

	// ErrNoPrevColumn indicates that the task is already in the first column
	ErrNoPrevColumn = errors.New("task is already in the first column")

	// ErrAlreadyFirstColumn indicates an attempt to move to previous when already at first
	ErrAlreadyFirstColumn = errors.New("already at first column")

	// ErrAlreadyLastColumn indicates an attempt to move to next when already at last
	ErrAlreadyLastColumn = errors.New("already at last column")
)
