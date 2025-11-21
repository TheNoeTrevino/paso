package models

// Label represents a tag that can be applied to tasks
type Label struct {
	ID    int
	Name  string
	Color string // Hex color code (e.g., "#7D56F4")
}
