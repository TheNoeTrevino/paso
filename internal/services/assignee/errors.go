package assignee

import "errors"

var (
	// ErrAssigneeNotFound is returned when an assignee is not found
	ErrAssigneeNotFound = errors.New("assignee not found")

	// ErrInvalidAssigneeID is returned when an assignee ID is invalid
	ErrInvalidAssigneeID = errors.New("invalid assignee ID")

	// ErrInvalidAssigneeName is returned when an assignee name is invalid
	ErrInvalidAssigneeName = errors.New("invalid assignee name")
)
