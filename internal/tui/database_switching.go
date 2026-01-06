package tui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
)

// tickConnectingSpinner returns a command that sends SpinnerTickMsg at intervals
// This creates a continuous animation loop while connecting
func tickConnectingSpinner() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return SpinnerTickMsg(t)
	})
}

// switchToDatabase attempts to connect to a new database and reload all data
// It can be called with either a DatabaseConfig or legacy string parameters
func (m Model) switchToDatabase(connStr string, dbName string) tea.Cmd {
	// Auto-detect database type based on connection string
	dbType := "sqlite"
	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		dbType = "postgres"
	}

	return m.switchToDatabaseConfig(config.DatabaseConfig{
		Name:             dbName,
		ConnectionString: connStr,
		Type:             dbType,
	})
}

// switchToDatabaseConfig attempts to connect using a DatabaseConfig
func (m Model) switchToDatabaseConfig(cfg config.DatabaseConfig) tea.Cmd {
	// Start the spinner animation loop
	// Note: StartConnecting() must be called by the caller in the Update handler
	// before calling this function, so the model change is visible immediately
	tickCmd := tickConnectingSpinner()

	return tea.Batch(tickCmd, func() tea.Msg {
		slog.Info("attempting to switch database", "name", cfg.Name)

		// Parse and validate connection string
		if cfg.ConnectionString == "" {
			err := fmt.Errorf("invalid connection string: empty")
			slog.Error("connection string validation failed", "error", err)
			return databaseConnectionError{err: err}
		}

		// Create context with timeout for connection
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Determine database type from connection string or config
		var dbConfig database.Config
		connStr := cfg.ConnectionString

		if connStr == "local" || connStr == "" {
			// SQLite local database
			slog.Info("using local SQLite database")
			home, err := os.UserHomeDir()
			if err != nil {
				slog.Error("failed to get home directory", "error", err)
				return databaseConnectionError{err: fmt.Errorf("failed to get home directory: %w", err)}
			}
			dbConfig = database.Config{
				Type:       database.SQLite,
				SQLitePath: filepath.Join(home, ".paso", "tasks.db"),
			}
		} else {
			// PostgreSQL remote database - validate and parse connection string
			slog.Info("using PostgreSQL remote database")
			slog.Debug("parsing PostgreSQL connection string")
			normalizedConnStr, err := database.ParseConnectionString(connStr)
			if err != nil {
				slog.Error("failed to parse connection string", "error", err)
				return databaseConnectionError{err: fmt.Errorf("invalid connection string: %w", err)}
			}

			// Parse the normalized connection string into config components
			dbConfig, err = database.ParsePostgresConnectionString(normalizedConnStr)
			if err != nil {
				slog.Error("failed to create database config from connection string", "error", err)
				return databaseConnectionError{err: fmt.Errorf("failed to parse connection string: %w", err)}
			}
		}

		// Test connection first
		slog.Debug("testing database connection", "type", dbConfig.Type)
		connStr, err := dbConfig.ConnectionString()
		if err != nil {
			slog.Error("failed to get connection string", "error", err, "database_name", cfg.Name)
			return databaseConnectionError{err: fmt.Errorf("failed to get connection string: %w", err)}
		}
		if err := database.TestConnection(connStr, dbConfig.Type); err != nil {
			slog.Error("connection test failed", "error", err, "database_name", cfg.Name)
			return databaseConnectionError{err: fmt.Errorf("connection test failed: %w", err)}
		}
		slog.Info("connection test passed")

		// Initialize new database connection
		slog.Info("initializing database connection", "database_name", cfg.Name)
		newDB, err := database.InitDB(ctx, dbConfig, cfg.Name)
		if err != nil {
			slog.Error("failed to initialize database", "error", err, "database_name", cfg.Name)
			return databaseConnectionError{err: fmt.Errorf("failed to initialize database: %w", err)}
		}
		slog.Info("database initialized successfully", "database_name", cfg.Name)

		// Create new app with new database connection
		// NOTE: Do NOT close old database or modify m.DB/m.App here!
		// The swap happens synchronously in the message handler to avoid race conditions.
		newApp, err := app.New(newDB,
			app.WithDatabaseType(dbConfig.Type),
			app.WithEventPublisher(m.EventClient),
		)
		if err != nil {
			slog.Error("failed to initialize application", "error", err, "database_name", cfg.Name)
			return databaseConnectionError{err: fmt.Errorf("failed to initialize application: %w", err)}
		}

		return databaseConnected{
			dbType: dbConfig.Type,
			dbName: cfg.Name,
			newDB:  newDB,
			newApp: newApp,
		}
	})
}

// handleDisconnectDatabase switches back to local SQLite
func (m Model) handleDisconnectDatabase() (tea.Model, tea.Cmd) {
	// Use "local" as the connection string - switchToDatabaseConfig handles this specially
	m.DatabasePicker.StartConnecting("Local")
	return m, m.switchToDatabase("local", "Local")
}

// reloadAllData reloads projects, columns, tasks, and labels after database switch
func (m Model) reloadAllData() tea.Cmd {
	return func() tea.Msg {
		slog.Info("reloadAllData closure executing", "m.App", m.App, "m.App.ColumnService", m.App.ColumnService)
		ctx, cancel := m.DBContext()
		defer cancel()

		// Load all projects
		projects, err := m.App.ProjectService.GetAllProjects(ctx)
		if err != nil {
			return databaseConnectionError{err: fmt.Errorf("failed to reload projects: %w", err)}
		}

		// Determine which project to load data for
		currentProjectID := 0
		if len(projects) > 0 {
			// Try to use current project if available in AppState
			if currentProject := m.AppState.GetCurrentProject(); currentProject != nil {
				// Check if this project still exists in the new database
				for _, p := range projects {
					if p.ID == currentProject.ID {
						currentProjectID = currentProject.ID
						break
					}
				}
			}
			// If current project doesn't exist in new database, use first project
			if currentProjectID == 0 {
				currentProjectID = projects[0].ID
			}
		}

		// If we have no projects, return empty data
		if currentProjectID == 0 {
			return dataReloaded{
				projects: projects,
				columns:  []*models.Column{},
				tasks:    make(map[int][]*models.TaskSummary),
				labels:   []*models.Label{},
			}
		}

		// Load columns for current project
		columns, err := m.App.ColumnService.GetColumnsByProject(ctx, currentProjectID)
		if err != nil {
			return databaseConnectionError{err: fmt.Errorf("failed to reload columns: %w", err)}
		}

		// Load task summaries for current project
		tasks, err := m.App.TaskService.GetTaskSummariesByProject(ctx, currentProjectID)
		if err != nil {
			return databaseConnectionError{err: fmt.Errorf("failed to reload tasks: %w", err)}
		}

		// Load labels for current project
		labels, err := m.App.LabelService.GetLabelsByProject(ctx, currentProjectID)
		if err != nil {
			return databaseConnectionError{err: fmt.Errorf("failed to reload labels: %w", err)}
		}

		return dataReloaded{
			projects: projects,
			columns:  columns,
			tasks:    tasks,
			labels:   labels,
		}
	}
}
