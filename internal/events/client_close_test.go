package events

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

// ============================================================================

func TestClose_BeforeConnect(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Close before connecting should not error
	if err := client.Close(); err != nil {
		t.Errorf("Expected Close to succeed, got error: %v", err)
	}

	t.Logf("✓ Close before connect succeeds")
}

func TestClose_AfterConnect(t *testing.T) {
	socketPath, listener, _ := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Close after connecting
	if err := client.Close(); err != nil {
		t.Errorf("Expected Close to succeed, got error: %v", err)
	}

	// Verify connection is closed
	client.mu.Lock()
	connected := client.conn != nil
	client.mu.Unlock()

	if connected {
		t.Error("Expected connection to be closed")
	}

	t.Logf("✓ Close after connect succeeds")
}

func TestClose_Idempotent(t *testing.T) {
	socketPath, listener, _ := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Close multiple times
	if err := client.Close(); err != nil {
		t.Errorf("First close failed: %v", err)
	}

	if err := client.Close(); err != nil {
		t.Errorf("Second close should be idempotent, got error: %v", err)
	}

	t.Logf("✓ Close is idempotent")
}

// ============================================================================
// Notify Callback Tests
// ============================================================================

func TestSetNotifyFunc(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Set notify function
	var capturedLevel, capturedMessage string
	client.SetNotifyFunc(func(level, message string) {
		capturedLevel = level
		capturedMessage = message
	})

	// Trigger notification by calling notify directly
	client.notify("info", "test notification")

	if capturedLevel != "info" {
		t.Errorf("Expected level 'info', got '%s'", capturedLevel)
	}

	if capturedMessage != "test notification" {
		t.Errorf("Expected message 'test notification', got '%s'", capturedMessage)
	}

	t.Logf("✓ SetNotifyFunc works correctly")
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestConnect_InvalidSocketPath(t *testing.T) {
	// Use a path that's too long or invalid
	invalidPath := fmt.Sprintf("/tmp/%s.sock", string(make([]byte, 200)))

	client, err := NewClient(invalidPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err = client.Connect(ctx)
	if err == nil {
		t.Error("Expected Connect to fail with invalid socket path")
	}

	t.Logf("✓ Connect handles invalid socket path")
}
