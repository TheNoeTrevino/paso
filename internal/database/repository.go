package database

import (
	"database/sql"

	"github.com/thenoetrevino/paso/internal/events"
)

// Repository provides a unified interface to all data operations.
// It composes domain-specific repositories using struct embedding.
// The embedded struct methods are promoted and automatically satisfy the DataStore interface.
type Repository struct {
	*ProjectRepo
	*ColumnRepo
	*TaskRepo
	*LabelRepo
	eventClient events.EventPublisher
}

// NewRepository creates a new Repository instance wrapping the given database connection.
// eventClient is optional and may be nil if the daemon is not running.
func NewRepository(db *sql.DB, eventClient events.EventPublisher) *Repository {
	return &Repository{
		ProjectRepo: &ProjectRepo{db: db, eventClient: eventClient},
		ColumnRepo:  &ColumnRepo{db: db, eventClient: eventClient},
		TaskRepo:    &TaskRepo{db: db, eventClient: eventClient},
		LabelRepo:   &LabelRepo{db: db, eventClient: eventClient},
		eventClient: eventClient,
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
