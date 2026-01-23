package taskevent

import "errors"

var (
	ErrEmptyContent  = errors.New("event content cannot be empty")
	ErrInvalidTaskID = errors.New("task ID must be positive")
)
