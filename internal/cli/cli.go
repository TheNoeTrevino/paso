// Package cli provides the command-line interface context and initialization
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/logging"
	"github.com/thenoetrevino/paso/internal/testutil"
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
	if testAppVal := ctx.Value(testutil.TestAppKey); testAppVal != nil {
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

	// Initialize database
	db, err := database.InitDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Try to connect to daemon (optional - silent fallback)
	home, _ := os.UserHomeDir()
	socketPath := filepath.Join(home, ".paso", "paso.sock")

	var eventClient events.EventPublisher
	client, err := events.NewClient(socketPath)
	if err == nil {
		// Try to connect - if it fails, daemon isn't running (graceful degradation)
		if err := client.Connect(ctx); err == nil {
			eventClient = client
		}
	}

	application := app.New(db, eventClient)

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
