package cli

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
)

// Context holds resources needed for CLI commands
type Context struct {
	DB           *sql.DB
	EventClient  *events.Client
	IsConnected  bool
}

// InitContext initializes the CLI context with database and optional daemon connection
func InitContext(ctx context.Context) (*Context, error) {
	// Initialize database
	db, err := database.InitDB(ctx)
	if err != nil {
		return nil, err
	}

	// Try to connect to daemon (optional - don't fail if not available)
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: failed to get home directory for daemon connection: %v", err)
		return &Context{
			DB:          db,
			EventClient: nil,
			IsConnected: false,
		}, nil
	}

	socketPath := filepath.Join(home, ".paso", "paso.sock")
	eventClient, err := events.NewClient(socketPath)
	if err != nil {
		log.Printf("Warning: failed to create event client: %v", err)
		return &Context{
			DB:          db,
			EventClient: nil,
			IsConnected: false,
		}, nil
	}

	// Try to connect to daemon - this is optional
	if err := eventClient.Connect(ctx); err != nil {
		// Daemon not running - this is OK
		return &Context{
			DB:          db,
			EventClient: nil,
			IsConnected: false,
		}, nil
	}

	// Successfully connected to daemon
	return &Context{
		DB:          db,
		EventClient: eventClient,
		IsConnected: true,
	}, nil
}

// Close cleans up CLI context resources
func (c *Context) Close() error {
	var firstErr error

	if c.EventClient != nil {
		if err := c.EventClient.Close(); err != nil {
			firstErr = err
		}
	}

	if c.DB != nil {
		if err := c.DB.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
