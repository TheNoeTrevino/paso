package daemon

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thenoetrevino/paso/internal/events"
)

func getTestSocketPath(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	return filepath.Join(tmpDir, "test-paso.sock")
}

func setupTestDaemon(t *testing.T) (*Server, string) {
	t.Helper()
	socketPath := getTestSocketPath(t)

	server, err := NewServer(socketPath)
	if err != nil {
		t.Fatalf("Failed to create test daemon: %v", err)
	}

	t.Cleanup(func() {
		_ = server.Shutdown()
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() { _ = server.Start(ctx) }()

	// Wait for socket
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			time.Sleep(10 * time.Millisecond)
			return server, socketPath
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("Timeout waiting for daemon socket")
	return nil, ""
}

func connectRawClient(t *testing.T, socketPath string) (net.Conn, *json.Encoder, *json.Decoder) {
	t.Helper()

	conn, err := (&net.Dialer{}).DialContext(context.Background(), "unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
	})

	return conn, json.NewEncoder(conn), json.NewDecoder(conn)
}

func sendSubscribeMessage(t *testing.T, encoder *json.Encoder, projectID int) {
	t.Helper()
	msg := events.Message{
		Version:   events.ProtocolVersion,
		Type:      "subscribe",
		Subscribe: &events.SubscribeMessage{ProjectID: projectID},
	}
	if err := encoder.Encode(msg); err != nil {
		t.Fatalf("Failed to send subscribe: %v", err)
	}
}

func waitForEvent(t *testing.T, ch <-chan events.Event, timeout time.Duration) events.Event {
	t.Helper()
	select {
	case event, ok := <-ch:
		if !ok {
			t.Fatal("Channel closed")
		}
		return event
	case <-time.After(timeout):
		t.Fatalf("Timeout waiting for event")
		return events.Event{}
	}
}

func waitForNoEvent(t *testing.T, ch <-chan events.Event, timeout time.Duration) {
	t.Helper()
	select {
	case event := <-ch:
		t.Fatalf("Unexpected event: %+v", event)
	case <-time.After(timeout):
		// Success
	}
}

func setupTestClient(t *testing.T, socketPath string) *events.Client {
	t.Helper()
	client, err := events.NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Cleanup(func() {
		_ = client.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	return client
}

func logServerState(t *testing.T, server *Server, label string) {
	t.Helper()
	t.Logf("=== Server State: %s ===", label)
	t.Logf("  Server: %p", server)
	t.Logf("========================")
}
