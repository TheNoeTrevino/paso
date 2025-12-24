package launcher

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/logging"
	"github.com/thenoetrevino/paso/internal/tui/core"
)

// Launch starts the TUI application
func Launch() error {
	// Initialize logging to file before anything else
	if err := logging.Init(); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	// Create root context with signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
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
		slog.Warn("failed to create daemon client", "message", daemonErr.Message, "hint", daemonErr.Hint)
		slog.Info("continuing without live updates")
		eventClient = nil
	} else {
		// Try to connect to daemon
		if err := eventClient.Connect(ctx); err != nil {
			daemonErr := events.ClassifyDaemonError(err)
			slog.Warn("failed to connect to daemon", "message", daemonErr.Message, "hint", daemonErr.Hint)
			slog.Info("continuing without live updates")
			eventClient = nil
		}
	}

	// Cleanup daemon connection on exit
	defer func() {
		if eventClient != nil {
			if err := eventClient.Close(); err != nil {
				slog.Error("error closing event client", "error", err)
			}
		}
	}()

	initCtx := context.Background()
	db, err := database.InitDB(initCtx)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// database cleanup
	defer func() {
		// Create drain context with 5-second timeout
		drainCtx, drainCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer drainCancel()

		// Allow time for in-flight operations to complete
		select {
		case <-drainCtx.Done():
			slog.Info("drain period complete, closing database")
		case <-time.After(100 * time.Millisecond):
			// Small delay to allow operations to wrap up
		}

		if err := db.Close(); err != nil {
			slog.Error("error closing database", "error", err)
		}
	}()

	repo := database.NewRepository(db, eventClient)
	application := app.New(repo, eventClient)
	tuiApp := core.New(ctx, application, cfg, eventClient)
	p := tea.NewProgram(tuiApp, tea.WithContext(ctx))

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
			return fmt.Errorf("error running program: %w", err)
		}
	case <-ctx.Done():
		slog.Info("shutdown signal received, cleaning up")
		// Give the program 5 seconds to clean up database queries still running
		time.Sleep(5 * time.Second)
	}

	return nil
}
