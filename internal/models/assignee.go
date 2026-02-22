package models

import (
	"fmt"
	"strconv"
	"time"
)

// Assignee represents a lightweight user identity for task assignment
// Assignees are global to the database, not project-scoped
type Assignee struct {
	ID        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (a *Assignee) PickerLabel() string {
	return fmt.Sprintf("%s (#%d)", a.Name, a.ID)
}

func (a *Assignee) PickerValue() string {
	return strconv.Itoa(a.ID)
}
