package models

import (
	"fmt"
	"strconv"
)

// Column represents a kanban board column (e.g., "Todo", "In Progress", "Done")
// Columns are organized as a doubly-linked list using PrevID and NextID pointers
// Each column belongs to a specific project
type Column struct {
	ID                   int    // Unique identifier for the column
	Name                 string // Display name of the column
	ProjectID            int    // ID of the project this column belongs to
	PrevID               *int   // ID of the previous column (NULL for head)
	NextID               *int   // ID of the next column (NULL for tail)
	HoldsReadyTasks      bool   // Whether tasks in this column are considered "ready" for work
	HoldsCompletedTasks  bool   // Whether tasks in this column are considered "completed"
	HoldsInProgressTasks bool   // Whether tasks in this column are considered "in progress"
}

func (c *Column) PickerLabel() string {
	return fmt.Sprintf("%s (#%d)", c.Name, c.ID)
}

func (c *Column) PickerValue() string {
	return strconv.Itoa(c.ID)
}
