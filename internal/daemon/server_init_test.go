package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewServer_Success(t *testing.T) {
	socketPath := getTestSocketPath(t)

	server, err := NewServer(socketPath)
	if err != nil {
		t.Fatalf("Expected NewServer to succeed, got error: %v", err)
	}
	defer func() { _ = server.Shutdown() }()

	// Verify socket file was created
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Error("Expected socket file to be created")
	}

	// Verify server fields initialized
	if server == nil {
		t.Fatal("Expected server to be non-nil")
	}

	t.Logf("✓ Server created successfully at %s", socketPath)
}

func TestNewServer_DirectoryCreation(t *testing.T) {
	// Use t.TempDir() which ensures cleanup
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "subdirs", "paso.sock")

	server, err := NewServer(nestedPath)
	if err != nil {
		t.Fatalf("Expected NewServer to create nested directories, got error: %v", err)
	}
	defer func() { _ = server.Shutdown() }()

	// Verify directories were created
	dir := filepath.Dir(nestedPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Expected directory %s to be created", dir)
	}

	// Verify socket file exists
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("Expected socket file to be created in nested directory")
	}

	t.Logf("✓ Nested directories created successfully: %s", nestedPath)
}

func TestNewServer_StaleSocketCleanup(t *testing.T) {
	socketPath := getTestSocketPath(t)

	// Create a stale socket file
	f, err := os.Create(socketPath)
	if err != nil {
		t.Fatalf("Failed to create stale socket file: %v", err)
	}
	_ = f.Close()

	// Verify stale socket exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Fatal("Stale socket file should exist before NewServer")
	}

	// Create new server (should remove stale socket)
	server, err := NewServer(socketPath)
	if err != nil {
		t.Fatalf("Expected NewServer to succeed after removing stale socket, got error: %v", err)
	}
	defer func() { _ = server.Shutdown() }()

	// Verify new socket was created (the old one was removed)
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Error("Expected new socket file to be created")
	}

	t.Logf("✓ Stale socket cleaned up successfully")
}

func TestNewServer_EnvVarConfiguration(t *testing.T) {
	t.Setenv("PASO_DAEMON_BROADCAST_BUFFER", "200")
	t.Setenv("PASO_DAEMON_CLIENT_BUFFER", "20")

	socketPath := getTestSocketPath(t)
	server, err := NewServer(socketPath)
	if err != nil {
		t.Fatalf("Expected NewServer to succeed, got error: %v", err)
	}
	defer func() { _ = server.Shutdown() }()

	// Note: We can't directly verify buffer sizes since they're unexported
	// But we can verify the server was created successfully with env vars set
	t.Logf("✓ Server created with custom buffer sizes from env vars")
}
