package logging

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

// Logger is the global slog instance for the application
var Logger *slog.Logger

// Init initializes the logging system, writing logs to ~/.paso/logs/paso.log
// Uses text format for human readability.
func Init() error {
	// Create ~/.paso/logs/ directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	logDir := filepath.Join(homeDir, ".paso", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Open log file in append mode
	logPath := filepath.Join(logDir, "paso.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Create text handler (human readable)
	handler := slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	Logger = slog.New(handler)
	slog.SetDefault(Logger)

	// Redirect standard log package output (used by goose) to the same file
	log.SetOutput(file)
	log.SetFlags(log.LstdFlags) // Include timestamp

	return nil
}
