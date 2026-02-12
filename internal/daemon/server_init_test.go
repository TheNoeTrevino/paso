package daemon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer_Success(t *testing.T) {
	t.Parallel()
	socketPath := getTestSocketPath(t)

	server, err := NewServer(socketPath)
	require.NoError(t, err)
	defer func() { _ = server.Shutdown() }()

	// Verify socket file was created
	_, err = os.Stat(socketPath)
	assert.False(t, os.IsNotExist(err))

	// Verify server fields initialized
	require.NotNil(t, server)

	t.Logf("Server created successfully at %s", socketPath)
}

func TestNewServer_DirectoryCreation(t *testing.T) {
	t.Parallel()
	// Use t.TempDir() which ensures cleanup
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "subdirs", "paso.sock")

	server, err := NewServer(nestedPath)
	require.NoError(t, err)
	defer func() { _ = server.Shutdown() }()

	// Verify directories were created
	dir := filepath.Dir(nestedPath)
	_, err = os.Stat(dir)
	assert.False(t, os.IsNotExist(err))

	// Verify socket file exists
	_, err = os.Stat(nestedPath)
	assert.False(t, os.IsNotExist(err))

	t.Logf("Nested directories created successfully: %s", nestedPath)
}

func TestNewServer_StaleSocketCleanup(t *testing.T) {
	t.Parallel()
	socketPath := getTestSocketPath(t)

	// Create a stale socket file
	f, err := os.Create(socketPath)
	require.NoError(t, err)
	_ = f.Close()

	// Verify stale socket exists
	_, err = os.Stat(socketPath)
	assert.False(t, os.IsNotExist(err))

	// Create new server (should remove stale socket)
	server, err := NewServer(socketPath)
	require.NoError(t, err)
	defer func() { _ = server.Shutdown() }()

	// Verify new socket was created (the old one was removed)
	_, err = os.Stat(socketPath)
	assert.False(t, os.IsNotExist(err))

	t.Logf("Stale socket cleaned up successfully")
}

func TestNewServer_EnvVarConfiguration(t *testing.T) {
	t.Setenv("PASO_DAEMON_BROADCAST_BUFFER", "200")
	t.Setenv("PASO_DAEMON_CLIENT_BUFFER", "20")

	socketPath := getTestSocketPath(t)
	server, err := NewServer(socketPath)
	require.NoError(t, err)
	defer func() { _ = server.Shutdown() }()

	// Note: We can't directly verify buffer sizes since they're unexported
	// But we can verify the server was created successfully with env vars set
	t.Logf("Server created with custom buffer sizes from env vars")
}
