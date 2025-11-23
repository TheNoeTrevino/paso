package models

import "time"

// Project represents a container for kanban columns and tasks
// Projects are the top-level organizational unit in Paso
type Project struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
