package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
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

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to daemon for live updates (optional - daemon may not be running)
	home := os.Getenv("HOME")
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	socketPath := filepath.Join(home, ".paso", "paso.sock")

	eventClient, err := events.NewClient(socketPath)
	if err != nil {
		// Daemon may not be available, log warning but continue
		daemonErr := events.ClassifyDaemonError(err)
		log.Printf("Warning: Failed to create daemon client: %s", daemonErr.Message)
		log.Printf("Hint: %s", daemonErr.Hint)
		log.Println("Continuing without live updates...")
		eventClient = nil
	} else {
		// Try to connect to daemon
		if err := eventClient.Connect(ctx); err != nil {
			daemonErr := events.ClassifyDaemonError(err)
			log.Printf("Warning: Failed to connect to daemon: %s", daemonErr.Message)
			log.Printf("Hint: %s", daemonErr.Hint)
			log.Println("Continuing without live updates...")
			eventClient = nil
		}
	}

	// Cleanup daemon connection on exit
	defer func() {
		if eventClient != nil {
			eventClient.Close()
		}
	}()

	initCtx := context.Background()
	db, err := database.InitDB(initCtx)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// database cleanup
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

	repo := database.NewRepository(db, eventClient)
	model := tui.InitialModel(ctx, repo, cfg, eventClient)
	p := tea.NewProgram(model, tea.WithContext(ctx))

	// goroutine to monitor cancellation
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
		// Give the program 5 seconds to clean up database queties still running
		time.Sleep(5 * time.Second)
	}
}
