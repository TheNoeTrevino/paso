package models

// Column represents a kanban board column (e.g., "Todo", "In Progress", "Done")
// Columns are organized as a doubly-linked list using PrevID and NextID pointers
type Column struct {
	ID     int    // Unique identifier for the column
	Name   string // Display name of the column
	PrevID *int   // ID of the previous column (NULL for head)
	NextID *int   // ID of the next column (NULL for tail)
}
