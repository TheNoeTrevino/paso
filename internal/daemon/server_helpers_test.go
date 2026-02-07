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

// subscriptionWaiter allows tests to install the OnSubscribe callback before
// sending subscribe messages, then wait for those subscriptions to be processed.
// This avoids the race where a subscribe message is processed before the
// callback is installed.
type subscriptionWaiter struct {
	t      *testing.T
	server *Server
	ch     chan int
	count  int
}

// expectSubscriptions installs the OnSubscribe callback and returns a waiter.
// Call this before sending subscribe messages, then call waiter.Wait() after.
// Subscriptions to projectID=0 (the auto-subscribe from Client.Connect) are
// ignored since tests always subscribe to non-zero project IDs.
func expectSubscriptions(t *testing.T, server *Server, count int) *subscriptionWaiter {
	t.Helper()
	ch := make(chan int, count+10)
	server.onSubscribeMu.Lock()
	server.OnSubscribe = func(projectID int) {
		if projectID != 0 {
			ch <- projectID
		}
	}
	server.onSubscribeMu.Unlock()
	return &subscriptionWaiter{t: t, server: server, ch: ch, count: count}
}

func (w *subscriptionWaiter) Wait() {
	w.t.Helper()
	timeout := time.After(2 * time.Second)
	received := 0

	for received < w.count {
		select {
		case <-w.ch:
			received++
		case <-timeout:
			require.Fail(w.t, "Timeout waiting for subscriptions", "expected %d, got %d", w.count, received)
		}
	}

	w.server.onSubscribeMu.Lock()
	w.server.OnSubscribe = nil
	w.server.onSubscribeMu.Unlock()
}
