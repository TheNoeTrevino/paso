package models

import "time"

// StandupLog represents an immutable standup log entry for a project.
type StandupLog struct {
	ID        int
	ProjectID int
	Content   string
	CreatedAt time.Time
}
