package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/tui"
)

func main() {
	// Create root context with signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database with non-cancellable context
	// Migrations must complete to avoid partial schema states
	initCtx := context.Background()
	db, err := database.InitDB(initCtx)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Database cleanup with graceful drain period
	defer func() {
		// Create drain context with 5-second timeout
		drainCtx, drainCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer drainCancel()

		// Allow time for in-flight operations to complete
		select {
		case <-drainCtx.Done():
			log.Println("Drain period complete, closing database")
		case <-time.After(100 * time.Millisecond):
			// Small delay to allow operations to wrap up
		}

		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Create repository wrapping the database
	repo := database.NewRepository(db)

	// Create initial TUI model with root context
	model := tui.InitialModel(repo, cfg)

	// Create Bubble Tea program with context support
	p := tea.NewProgram(model, tea.WithContext(ctx))

	// Run program in goroutine to monitor cancellation
	errChan := make(chan error, 1)
	go func() {
		_, err := p.Run()
		errChan <- err
	}()

	// Wait for program completion or cancellation
	select {
	case err := <-errChan:
		if err != nil {
			fmt.Printf("Error running program: %v\n", err)
			log.Fatal(err)
		}
	case <-ctx.Done():
		log.Println("Shutdown signal received, cleaning up...")
		// Give the program 5 seconds to clean up
		time.Sleep(5 * time.Second)
	}
}
