package models

// Priority represents a task priority level
type Priority struct {
	ID          int
	Description string
	Color       string
}

// Type represents a task type
type Type struct {
	ID          int
	Description string
}
