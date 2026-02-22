package standuplog

import "errors"

var (
	ErrEmptyContent     = errors.New("standup log content cannot be empty")
	ErrInvalidProjectID = errors.New("project ID must be positive")
	ErrInvalidLogID     = errors.New("standup log ID must be positive")
)
