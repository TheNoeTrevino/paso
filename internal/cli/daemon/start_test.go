package daemon

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	internaldaemon "github.com/thenoetrevino/paso/internal/daemon"
)

// TestRunDaemon_CreatesSocketAndShutdown exercises the same code path as
// runDaemon() — resolve paso dir, create it, start the daemon server, then
// verify it accepts connections and shuts down cleanly on context cancellation.
//
// We can't call runDaemon() directly because it creates its own
// signal.NotifyContext that can only be cancelled via OS signals. Instead we
// replicate the key logic (path resolution -> mkdir -> NewServer -> Start)
// using a controlled HOME so the test is deterministic.
func TestRunDaemon_CreatesSocketAndShutdown(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	pasoDir, err := getPasoDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(fakeHome, ".paso"), pasoDir)

	socketPath, err := getSocketPath()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(fakeHome, ".paso", "paso.sock"), socketPath)

	// Create the .paso directory (mirrors runDaemon)
	require.NoError(t, os.MkdirAll(pasoDir, 0700))

	info, err := os.Stat(pasoDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
	assert.Equal(t, os.FileMode(0700), info.Mode().Perm())

	// Start the daemon server with a cancellable context
	server, err := internaldaemon.NewServer(socketPath)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.Start(ctx)
	}()

	// Wait for socket to appear
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, statErr := os.Stat(socketPath); statErr == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	_, err = os.Stat(socketPath)
	require.NoError(t, err, "daemon socket was never created")

	// Verify we can connect
	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.DialContext(ctx, "unix", socketPath)
	require.NoError(t, err, "could not connect to daemon socket")
	_ = conn.Close()

	// Cancel context to trigger shutdown
	cancel()

	select {
	case serverErr := <-serverDone:
		assert.NoError(t, serverErr, "server should shut down without error")
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}
