// Package cli provides the command-line interface context and initialization
package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/appcontext"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/logging"
	"github.com/thenoetrevino/paso/internal/user"
)

// CLI represents the CLI application context
type CLI struct {
	App         *app.App // Application container with services
	eventClient events.EventPublisher
	ctx         context.Context
}

// GetCLIFromContext retrieves a CLI instance from context (for testing)
// If not found, creates a new CLI instance
func GetCLIFromContext(ctx context.Context) (*CLI, error) {
	// Check if testApp is in context (test mode)
	if testAppVal := ctx.Value(appcontext.AppKey); testAppVal != nil {
		if testApp, ok := testAppVal.(*app.App); ok {
			return &CLI{
				App: testApp,
				ctx: ctx,
			}, nil
		}
	}

	// Create a new CLI instance (production mode)
	return NewCLI(ctx)
}

// NewCLI initializes the CLI with database and optional daemon connection
func NewCLI(ctx context.Context) (*CLI, error) {
	return NewCLIWithApp(ctx, nil)
}

// NewCLIWithApp initializes the CLI with an optional pre-created app (for testing)
func NewCLIWithApp(ctx context.Context, testApp *app.App) (*CLI, error) {
	// If app is provided (test mode), use it directly
	if testApp != nil {
		return &CLI{
			App: testApp,
			ctx: ctx,
		}, nil
	}

	// Initialize logging to file before database initialization
	// This ensures goose migration logs go to the log file instead of stdout
	if err := logging.Init(); err != nil {
		// If logging fails, we can still continue - it's non-critical for CLI
		// but we won't suppress migration logs
		_ = err
	}

	// Load config to get active database
	cfg, _ := config.Load()

	// Determine database configuration
	var dbConfig database.Config
	var dbName string

	if activeDB := cfg.GetActiveDatabase(); activeDB != nil {
		dbName = activeDB.Name
		if activeDB.Type == "postgres" {
			dbConfig = database.Config{
				Type:            database.PostgreSQL,
				PostgresConnStr: activeDB.ConnectionString,
			}
		} else {
			connStr := activeDB.ConnectionString
			if strings.HasPrefix(connStr, "~/") {
				home, _ := os.UserHomeDir()
				connStr = filepath.Join(home, connStr[2:])
			}
			dbConfig = database.Config{
				Type:       database.SQLite,
				SQLitePath: connStr,
			}
		}
	} else {
		home, _ := os.UserHomeDir()
		dbPath := filepath.Join(home, ".paso", "tasks.db")
		dbConfig = database.Config{
			Type:       database.SQLite,
			SQLitePath: dbPath,
		}
		dbName = "Local"
	}

	db, err := database.InitDB(ctx, dbConfig, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Try to connect to daemon (optional - silent fallback)
	home, _ := os.UserHomeDir()
	socketPath := filepath.Join(home, ".paso", "paso.sock")

	var eventClient events.EventPublisher
	client, err := events.NewClient(socketPath)
	if err != nil {
		// Daemon socket not available - graceful degradation (eventClient remains nil)
	} else {
		err = client.Connect(ctx)
		if err != nil {
			// Connection failed - graceful degradation (eventClient remains nil)
		} else {
			eventClient = client
		}
	}

	// Build options for app initialization
	var appOpts []app.Option
	if eventClient != nil {
		appOpts = append(appOpts, app.WithEventPublisher(eventClient))
	}

	application, err := app.New(db, appOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize application: %w", err)
	}

	activeAssignee := cfg.GetActiveAssignee()
	if activeAssignee == "" {
		activeAssignee = user.GetCurrentUsername()
	}

	_, err = application.AssigneeService.GetOrCreate(ctx, activeAssignee)
	if err != nil {
		slog.Warn("failed to initialize active assignee", "assignee", activeAssignee, "error", err)
	}

	return &CLI{
		App:         application,
		eventClient: eventClient,
		ctx:         ctx,
	}, nil
}

// Close cleans up CLI resources
func (c *CLI) Close() error {
	if c.eventClient != nil {
		if err := c.eventClient.Close(); err != nil {
			// Log but don't fail - best effort cleanup
			// Still attempt to close the app
			_ = c.App.Close()
			return fmt.Errorf("failed to close event client: %w", err)
		}
	}
	return c.App.Close()
}
