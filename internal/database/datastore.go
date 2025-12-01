package database

// DataStore defines the unified interface for all data operations needed by the TUI.
// This interface is composed of smaller, domain-specific interfaces following the
// Interface Segregation Principle. Consumers can depend on smaller interfaces
// (e.g., TaskRepository, ColumnRepository) for better testability and clearer dependencies.
type DataStore interface {
	ProjectRepository
	ColumnRepository
	TaskRepository
	LabelRepository
}
