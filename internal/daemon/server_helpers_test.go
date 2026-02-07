package daemon

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

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

	require.Fail(t, "Timeout waiting for daemon socket")
	return nil, ""
}

func connectRawClient(t *testing.T, socketPath string) (net.Conn, *json.Encoder, *json.Decoder) {
	t.Helper()

	conn, err := (&net.Dialer{}).DialContext(context.Background(), "unix", socketPath)
	require.NoError(t, err)

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
	err := encoder.Encode(msg)
	require.NoError(t, err)
}

func waitForEvent(t *testing.T, ch <-chan events.Event, timeout time.Duration) events.Event {
	t.Helper()
	select {
	case event, ok := <-ch:
		require.True(t, ok, "Channel closed")
		return event
	case <-time.After(timeout):
		require.Fail(t, "Timeout waiting for event")
		return events.Event{}
	}
}

func waitForNoEvent(t *testing.T, ch <-chan events.Event, timeout time.Duration) {
	t.Helper()
	select {
	case event := <-ch:
		require.Fail(t, "Unexpected event", "%+v", event)
	case <-time.After(timeout):
		// Success
	}
}

func setupTestClient(t *testing.T, socketPath string) *events.Client {
	t.Helper()
	client, err := events.NewClient(socketPath)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = client.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)

	return client
}

func logServerState(t *testing.T, server *Server, label string) {
	t.Helper()
	t.Logf("=== Server State: %s ===", label)
	t.Logf("  Server: %p", server)
	t.Logf("========================")
}

func waitForSubscriptions(t *testing.T, server *Server, count int) {
	t.Helper()

	ch := make(chan int, count)
	server.OnSubscribe = func(projectID int) {
		ch <- projectID
	}

	timeout := time.After(2 * time.Second)
	received := 0

	for received < count {
		select {
		case <-ch:
			received++
		case <-timeout:
			require.Fail(t, "Timeout waiting for subscriptions", "expected %d, got %d", count, received)
		}
	}

	server.OnSubscribe = nil
}
