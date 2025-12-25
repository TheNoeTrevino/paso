package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
)

// CLI represents the CLI application context
type CLI struct {
	App         *app.App // Application container with services
	eventClient events.EventPublisher
	ctx         context.Context
}

// NewCLI initializes the CLI with database and optional daemon connection
func NewCLI(ctx context.Context) (*CLI, error) {
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
		c.eventClient.Close()
	}
	return c.App.Close()
}
