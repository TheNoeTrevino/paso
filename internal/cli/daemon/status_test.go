package daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestIsSocketReachable_RunningDaemon(t *testing.T) {
	t.Parallel()

	_, socketPath := fixtures.SetupTestDaemon(t)

	reachable := isSocketReachable(context.Background(), socketPath)

	assert.True(t, reachable)
}

func TestIsSocketReachable_NoSocket(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nonexistent.sock")

	reachable := isSocketReachable(context.Background(), path)

	assert.False(t, reachable)
}

func TestIsSocketReachable_StaleSocketFile(t *testing.T) {
	t.Parallel()

	socketPath := filepath.Join(t.TempDir(), "stale.sock")
	err := os.WriteFile(socketPath, []byte{}, 0600)
	require.NoError(t, err)

	reachable := isSocketReachable(context.Background(), socketPath)

	assert.False(t, reachable)
}

func TestIsSocketReachable_CancelledContext(t *testing.T) {
	t.Parallel()

	_, socketPath := fixtures.SetupTestDaemon(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	reachable := isSocketReachable(ctx, socketPath)

	assert.False(t, reachable)
}
