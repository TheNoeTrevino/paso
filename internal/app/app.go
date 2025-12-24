package app

import (
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
)

// App holds all application services and provides dependency injection.
// This is the main application container that manages service lifecycles.
type App struct {
	// Repository layer (direct database access)
	repo database.DataStore

	// Event system for live updates
	eventClient events.EventPublisher

	// Service layer (business logic)
	TaskService taskservice.Service
	// ProjectService projectservice.Service  // TODO: Implement in Step 1.7
	// ColumnService  columnservice.Service   // TODO: Implement in Step 1.7
	// LabelService   labelservice.Service    // TODO: Implement in Step 1.7
}

// New creates a new App with all services initialized.
// This is the single entry point for creating the application container.
func New(repo database.DataStore, eventClient events.EventPublisher) *App {
	return &App{
		repo:        repo,
		eventClient: eventClient,
		TaskService: taskservice.NewService(repo, eventClient),
		// ProjectService: projectservice.NewService(repo, eventClient),  // TODO
		// ColumnService:  columnservice.NewService(repo, eventClient),   // TODO
		// LabelService:   labelservice.NewService(repo, eventClient),    // TODO
	}
}

// Repo returns the underlying repository for direct database access.
// This is provided for gradual migration - eventually all access should go through services.
func (a *App) Repo() database.DataStore {
	return a.repo
}

// Close performs cleanup of application resources.
// Currently a no-op, but provided for future resource management needs.
func (a *App) Close() error {
	// Future: Close any service-specific resources
	return nil
}
