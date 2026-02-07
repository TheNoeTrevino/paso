package events

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestNewClient_Success(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Expected NewClient to succeed, got error: %v", err)
	}
	defer func() { _ = client.Close() }()

	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	if client.socketPath != socketPath {
		t.Errorf("Expected socket path %s, got %s", socketPath, client.socketPath)
	}

	if client.debounce == 0 {
		t.Error("Expected debounce duration to be set")
	}

	t.Logf("✓ Client created successfully with debounce: %v", client.debounce)
}

func TestNewClient_CustomDebounce(t *testing.T) {
	t.Setenv("PASO_EVENT_DEBOUNCE_MS", "250")

	socketPath := filepath.Join(t.TempDir(), "paso.sock")
	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	expectedDebounce := 250 * time.Millisecond
	if client.debounce != expectedDebounce {
		t.Errorf("Expected debounce %v, got %v", expectedDebounce, client.debounce)
	}

	t.Logf("✓ Custom debounce set correctly: %v", client.debounce)
}

func TestConnect_Success(t *testing.T) {
	socketPath, listener, _ := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Expected Connect to succeed, got error: %v", err)
	}

	// Verify connection is established
	client.mu.Lock()
	connected := client.conn != nil
	client.mu.Unlock()

	if !connected {
		t.Error("Expected client to be connected")
	}

	t.Logf("✓ Client connected successfully")
}

func TestConnect_NoServer(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "nonexistent.sock")

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err = client.Connect(ctx)
	if err == nil {
		t.Error("Expected Connect to fail when server doesn't exist")
	}

	t.Logf("✓ Connect correctly failed: %v", err)
}

func TestConnect_ContextTimeout(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "timeout.sock")

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = client.Connect(ctx)
	if err == nil {
		t.Error("Expected Connect to fail with cancelled context")
	}

	t.Logf("✓ Connect respects context cancellation")
}
