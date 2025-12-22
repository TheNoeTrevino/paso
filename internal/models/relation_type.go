package models

// RelationType represents a type of relationship between tasks
type RelationType struct {
	ID         int
	PToCLabel  string // Label from parent's perspective (e.g., "Blocked By", "Parent")
	CToPLabel  string // Label from child's perspective (e.g., "Blocker", "Child")
	Color      string // Hex color code
	IsBlocking bool   // Special flag for blocking relationships
}
