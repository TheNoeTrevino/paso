package models

import (
	"fmt"
	"strconv"
)

// Label represents a tag that can be applied to tasks
// Labels are project-specific, similar to GitHub labels
type Label struct {
	ID        int
	Name      string
	Color     string // Hex color code (e.g., "#7D56F4")
	ProjectID int    // ID of the project this label belongs to
}

func (l *Label) PickerLabel() string {
	if l.Color != "" {
		return fmt.Sprintf("%s (#%d) — %s", l.Name, l.ID, l.Color)
	}
	return fmt.Sprintf("%s (#%d)", l.Name, l.ID)
}

func (l *Label) PickerValue() string {
	return strconv.Itoa(l.ID)
}
