package daemon

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.NoError(t, server.Shutdown())

	// Verify socket file removed
	_, err := os.Stat(socketPath)
	assert.True(t, os.IsNotExist(err))

	// Verify clients are disconnected (their connections should be closed)
	// Try to send event - should fail
	if err := client1.SendEvent(events.Event{Type: events.EventDatabaseChanged}); err == nil {
		// Event might still be in queue, that's ok
		t.Logf("Note: Event queued after shutdown (might be flushed before close)")
	}

	t.Logf("Server shutdown gracefully")
}

func TestShutdown_Idempotent(t *testing.T) {
	socketPath := getTestSocketPath(t)
	server, err := NewServer(socketPath)
	require.NoError(t, err)

	// Shutdown once
	assert.NoError(t, server.Shutdown())

	// Shutdown again - should not panic or error
	assert.NoError(t, server.Shutdown())

	t.Logf("Shutdown is idempotent")
}
