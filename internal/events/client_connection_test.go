package events

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_Success(t *testing.T) {
	t.Parallel()
	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	require.NotNil(t, client)

	assert.Equal(t, socketPath, client.socketPath)

	assert.NotZero(t, client.debounce)

	t.Logf("Client created successfully with debounce: %v", client.debounce)
}

func TestNewClient_CustomDebounce(t *testing.T) {
	t.Setenv("PASO_EVENT_DEBOUNCE_MS", "250")

	socketPath := filepath.Join(t.TempDir(), "paso.sock")
	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	expectedDebounce := 250 * time.Millisecond
	assert.Equal(t, expectedDebounce, client.debounce)

	t.Logf("Custom debounce set correctly: %v", client.debounce)
}

func TestConnect_Success(t *testing.T) {
	t.Parallel()
	socketPath, listener, _ := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	require.NoError(t, client.Connect(ctx))

	// Verify connection is established
	client.mu.Lock()
	connected := client.conn != nil
	client.mu.Unlock()

	assert.True(t, connected)

	t.Logf("Client connected successfully")
}

func TestConnect_NoServer(t *testing.T) {
	t.Parallel()
	socketPath := filepath.Join(t.TempDir(), "nonexistent.sock")

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err = client.Connect(ctx)
	assert.Error(t, err)

	t.Logf("Connect correctly failed: %v", err)
}

func TestConnect_ContextTimeout(t *testing.T) {
	t.Parallel()
	socketPath := filepath.Join(t.TempDir(), "timeout.sock")

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = client.Connect(ctx)
	assert.Error(t, err)

	t.Logf("Connect respects context cancellation")
}
