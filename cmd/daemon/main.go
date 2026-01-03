package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/thenoetrevino/paso/internal/daemon"
)

func main() {
	// Set up signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	// Get home directory from environment (set by systemd)
	home := os.Getenv("HOME")
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			slog.Error("failed to get home directory", "error", err)
			os.Exit(1)
		}
	}

	// Construct paths
	pasoDir := filepath.Join(home, ".paso")
	socketPath := filepath.Join(pasoDir, "paso.sock")

	// Ensure .paso directory exists with secure permissions
	if err := os.MkdirAll(pasoDir, 0700); err != nil {
		slog.Error("failed to create .paso directory", "error", err)
		os.Exit(1)
	}

	// Create and start the daemon server
	server, err := daemon.NewServer(socketPath)
	if err != nil {
		slog.Error("failed to create daemon", "error", err)
		os.Exit(1)
	}

	slog.Info("paso daemon starting", "socket_path", socketPath, "pid", os.Getpid())

	// Start the daemon (blocks until shutdown)
	if err := server.Start(ctx); err != nil {
		slog.Error("daemon error", "error", err)
		os.Exit(1)
	}

	slog.Info("paso daemon shutting down gracefully")
}
