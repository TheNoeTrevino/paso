package models

import "time"

// Assignee represents a lightweight user identity for task assignment
// Assignees are global to the database, not project-scoped
type Assignee struct {
	ID        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
