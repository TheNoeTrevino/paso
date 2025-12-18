package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
)

// CLI represents the CLI application context
type CLI struct {
	Repo *database.Repository
	ctx  context.Context
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
	eventClient, _ := events.NewClient(socketPath)

	repo := database.NewRepository(db, eventClient)

	return &CLI{
		Repo: repo,
		ctx:  ctx,
	}, nil
}

// Close cleans up CLI resources
func (c *CLI) Close() error {
	// Repository cleanup if needed
	return nil
}
