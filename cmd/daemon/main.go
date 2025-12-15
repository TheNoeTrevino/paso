package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/thenoetrevino/paso/internal/daemon"
)

func main() {
	// Set up logging for systemd
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix("paso-daemon: ")

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
			log.Fatalf("Failed to get home directory: %v", err)
		}
	}

	// Construct paths
	pasoDir := filepath.Join(home, ".paso")
	socketPath := filepath.Join(pasoDir, "paso.sock")

	// Ensure .paso directory exists with secure permissions
	if err := os.MkdirAll(pasoDir, 0700); err != nil {
		log.Fatalf("Failed to create .paso directory: %v", err)
	}

	// Create and start the daemon server
	server, err := daemon.NewServer(socketPath)
	if err != nil {
		log.Fatalf("Failed to create daemon: %v", err)
	}

	log.Printf("Paso daemon starting on %s", socketPath)
	log.Printf("Process ID: %d", os.Getpid())

	// Start the daemon (blocks until shutdown)
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Daemon error: %v", err)
	}

	log.Println("Paso daemon shutting down gracefully")
}
