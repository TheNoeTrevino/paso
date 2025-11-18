package models

// Column represents a kanban board column (e.g., "Todo", "In Progress", "Done")
type Column struct {
	ID       int
	Name     string
	Position int
}
