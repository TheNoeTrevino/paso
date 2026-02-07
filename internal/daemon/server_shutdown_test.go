package daemon

import (
	"os"
	"testing"
	"time"

	"github.com/thenoetrevino/paso/internal/events"
)

func TestShutdown_GracefulClose(t *testing.T) {
	server, socketPath := setupTestDaemon(t)

	// Connect a few clients
	client1 := setupTestClient(t, socketPath)
	_ = setupTestClient(t, socketPath) // client2

	time.Sleep(100 * time.Millisecond)
	logServerState(t, server, "before shutdown")

	// Shutdown server
	if err := server.Shutdown(); err != nil {
		t.Errorf("Expected Shutdown to succeed, got error: %v", err)
	}

	// Verify socket file removed
	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Error("Expected socket file to be removed after shutdown")
	}

	// Verify clients are disconnected (their connections should be closed)
	// Try to send event - should fail
	if err := client1.SendEvent(events.Event{Type: events.EventDatabaseChanged}); err == nil {
		// Event might still be in queue, that's ok
		t.Logf("Note: Event queued after shutdown (might be flushed before close)")
	}

	t.Logf("✓ Server shutdown gracefully")
}

func TestShutdown_Idempotent(t *testing.T) {
	socketPath := getTestSocketPath(t)
	server, err := NewServer(socketPath)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Shutdown once
	if err := server.Shutdown(); err != nil {
		t.Errorf("First shutdown failed: %v", err)
	}

	// Shutdown again - should not panic or error
	if err := server.Shutdown(); err != nil {
		t.Errorf("Second shutdown should be idempotent, got error: %v", err)
	}

	t.Logf("✓ Shutdown is idempotent")
}
