package app

import (
	"database/sql"

	"github.com/thenoetrevino/paso/internal/events"
	columnservice "github.com/thenoetrevino/paso/internal/services/column"
	labelservice "github.com/thenoetrevino/paso/internal/services/label"
	projectservice "github.com/thenoetrevino/paso/internal/services/project"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
)

// App holds all application services and provides dependency injection.
// This is the main application container that manages service lifecycles.
type App struct {
	// Event system for live updates
	eventClient events.EventPublisher

	// Service layer (business logic) - ONLY public interface
	TaskService    taskservice.Service
	ProjectService projectservice.Service
	ColumnService  columnservice.Service
	LabelService   labelservice.Service
}

// New creates a new App with all services initialized.
// This is the single entry point for creating the application container.
// Services use SQLC directly - no repository layer needed.
// Use functional options to customize the App initialization.
func New(db *sql.DB, opts ...Option) *App {
	// Create default configuration
	cfg := &appConfig{
		eventClient: nil,
		logger:      nil,
	}

	// Apply provided options
	for _, opt := range opts {
		opt(cfg)
	}

	// Create services with database connection
	// Each service creates its own SQLC queries instance internally
	return &App{
		eventClient:    cfg.eventClient,
		TaskService:    taskservice.NewService(db, cfg.eventClient),
		ProjectService: projectservice.NewService(db, cfg.eventClient),
		ColumnService:  columnservice.NewService(db, cfg.eventClient),
		LabelService:   labelservice.NewService(db, cfg.eventClient),
	}
}

// Close performs cleanup of application resources.
// Currently a no-op, but provided for future resource management needs.
func (a *App) Close() error {
	// Future: Close any service-specific resources
	return nil
}
