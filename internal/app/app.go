package app

import (
	"database/sql"

	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
	assigneeservice "github.com/thenoetrevino/paso/internal/services/assignee"
	columnservice "github.com/thenoetrevino/paso/internal/services/column"
	labelservice "github.com/thenoetrevino/paso/internal/services/label"
	projectservice "github.com/thenoetrevino/paso/internal/services/project"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
	"github.com/thenoetrevino/paso/internal/services/taskevent"
)

// App holds all application services and provides dependency injection.
// This is the main application container that manages service lifecycles.
type App struct {
	// Event system for live updates
	eventClient events.EventPublisher

	// Database configuration
	dbType database.DatabaseType

	// Service layer (business logic) - ONLY public interface
	TaskService     taskservice.Service
	ProjectService  projectservice.Service
	ColumnService   columnservice.Service
	LabelService    labelservice.Service
	AssigneeService assigneeservice.Service
}

// New creates a new App with all services initialized.
// This is the single entry point for creating the application container.
// Services use SQLC directly - no repository layer needed.
// Use functional options to customize the App initialization.
// Returns an error if any service fails to initialize (e.g., invalid database type).
func New(db *sql.DB, opts ...Option) (*App, error) {
	// Create default configuration
	cfg := &appConfig{
		eventClient: nil,
		logger:      nil,
		dbType:      database.SQLite, // Default to SQLite
	}

	// Apply provided options
	for _, opt := range opts {
		opt(cfg)
	}

	// Create services with database connection and type
	// Each service uses the database.Querier interface for database abstraction
	taskEventSvc, err := taskevent.NewService(db, cfg.dbType)
	if err != nil {
		return nil, err
	}

	taskSvc, err := taskservice.NewService(db, cfg.dbType, cfg.eventClient, taskEventSvc)
	if err != nil {
		return nil, err
	}

	projectSvc, err := projectservice.NewService(db, cfg.dbType, cfg.eventClient, nil)
	if err != nil {
		return nil, err
	}

	columnSvc, err := columnservice.NewService(db, cfg.dbType, cfg.eventClient)
	if err != nil {
		return nil, err
	}

	labelSvc, err := labelservice.NewService(db, cfg.dbType, cfg.eventClient)
	if err != nil {
		return nil, err
	}

	assigneeSvc, err := assigneeservice.NewService(db, cfg.dbType)
	if err != nil {
		return nil, err
	}

	return &App{
		eventClient:     cfg.eventClient,
		dbType:          cfg.dbType,
		TaskService:     taskSvc,
		ProjectService:  projectSvc,
		ColumnService:   columnSvc,
		LabelService:    labelSvc,
		AssigneeService: assigneeSvc,
	}, nil
}

// Close performs cleanup of application resources.
// Currently a no-op, but provided for future resource management needs.
func (a *App) Close() error {
	// Future: Close any service-specific resources
	return nil
}
