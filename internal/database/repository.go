package database

import (
	"database/sql"
)

// Repository provides a unified interface to all data operations.
// It composes domain-specific repositories using struct embedding.
// The embedded struct methods are promoted and automatically satisfy the DataStore interface.
type Repository struct {
	*ProjectRepo
	*ColumnRepo
	*TaskRepo
	*LabelRepo
}

// NewRepository creates a new Repository instance wrapping the given database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		ProjectRepo: &ProjectRepo{db: db},
		ColumnRepo:  &ColumnRepo{db: db},
		TaskRepo:    &TaskRepo{db: db},
		LabelRepo:   &LabelRepo{db: db},
	}
}

// Compile-time verification that Repository implements all interfaces
var (
	_ DataStore         = (*Repository)(nil)
	_ ProjectRepository = (*Repository)(nil)
	_ ColumnRepository  = (*Repository)(nil)
	_ TaskRepository    = (*Repository)(nil)
	_ LabelRepository   = (*Repository)(nil)
)
