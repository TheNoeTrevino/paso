package events

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClose_BeforeConnect(t *testing.T) {
	t.Parallel()
	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	require.NoError(t, err)

	// Close before connecting should not error
	err = client.Close()
	assert.NoError(t, err)

	t.Logf("Close before connect succeeds")
}

func TestClose_AfterConnect(t *testing.T) {
	t.Parallel()
	socketPath, listener, _ := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)

	// Close after connecting
	err = client.Close()
	assert.NoError(t, err)

	// Verify connection is closed
	client.mu.Lock()
	connected := client.conn != nil
	client.mu.Unlock()

	assert.False(t, connected)

	t.Logf("Close after connect succeeds")
}

func TestClose_Idempotent(t *testing.T) {
	t.Parallel()
	socketPath, listener, _ := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)

	// Close multiple times
	err = client.Close()
	assert.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)

	t.Logf("Close is idempotent")
}

func TestSetNotifyFunc(t *testing.T) {
	t.Parallel()
	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	// Set notify function
	var capturedLevel, capturedMessage string
	client.SetNotifyFunc(func(level, message string) {
		capturedLevel = level
		capturedMessage = message
	})

	// Trigger notification by calling notify directly
	client.notify("info", "test notification")

	assert.Equal(t, "info", capturedLevel)
	assert.Equal(t, "test notification", capturedMessage)

	t.Logf("SetNotifyFunc works correctly")
}

func TestConnect_InvalidSocketPath(t *testing.T) {
	t.Parallel()
	// Use a path that's too long or invalid
	invalidPath := fmt.Sprintf("/tmp/%s.sock", string(make([]byte, 200)))

	client, err := NewClient(invalidPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err = client.Connect(ctx)
	assert.Error(t, err)

	t.Logf("Connect handles invalid socket path")
}
