package events

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_DetectsDaemonDisconnect verifies that the client properly detects
// when the daemon socket closes and handles the disconnection gracefully.
//
// This test validates:
// - Client successfully connects to daemon initially
// - When daemon stops and socket closes, client detects disconnection
// - The readEvents function detects EOF from closed connection
// - Reconnection is attempted but fails (daemon not running)
// - Listen loop exits cleanly after failed reconnection
func TestClient_DetectsDaemonDisconnect(t *testing.T) {
	t.Parallel()

	// Setup controllable mock daemon
	socketPath, startFunc, stopFunc, messages := setupMockDaemonWithControl(t)

	// Start the daemon
	err := startFunc()
	require.NoError(t, err)
	t.Logf("Mock daemon started")

	// Create client
	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	// Set very short retry parameters for faster testing
	client.baseDelay = 50 * time.Millisecond
	client.maxRetries = 2
	client.setBatcherShutdownTimeoutForTest(25 * time.Millisecond)

	// Connect
	connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Connect(connectCtx)
	require.NoError(t, err)
	t.Logf("Client connected")

	// Drain initial subscribe
	select {
	case <-messages:
		t.Logf("Initial subscribe received")
	case <-time.After(1 * time.Second):
		require.Fail(t, "Timeout waiting for subscribe")
	}

	// Start Listen with client's internal context
	eventChan, err := client.Listen(client.ctx)
	require.NoError(t, err)

	// Monitor for channel close
	done := make(chan struct{})
	go func() {
		for range eventChan {
		}
		close(done)
	}()

	// Wait a bit for Listen to start reading
	time.Sleep(100 * time.Millisecond)

	// Simulate daemon crash: stop daemon and remove socket
	t.Logf("Simulating daemon disconnect...")
	start := time.Now()
	stopFunc()
	// Remove socket to fully simulate daemon crash (may already be removed by listener close)
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		require.NoError(t, err)
	}

	// Wait for Listen loop to detect disconnect and exit
	// With baseDelay=50ms and maxRetries=2: ~150ms for retries
	// Plus time to detect EOF (should be immediate when decoder reads)
	select {
	case <-done:
		elapsed := time.Since(start)
		t.Logf("Client detected disconnect in %v", elapsed)

		// Generous timeout: should complete in < 1 second
		if elapsed > 1*time.Second {
			t.Logf("Detection took longer than expected")
		}

	case <-time.After(10 * time.Second):
		require.Fail(t, "Client did not detect disconnect within 10s")
	}

	t.Logf("Test completed: disconnect detection verified")
}

// TestClient_ReconnectsAfterDaemonRestart verifies that the client successfully
// reconnects when the daemon comes back online after being stopped.
//
// This test validates:
// - Client connects to daemon initially
// - Client detects when daemon stops
// - Client triggers automatic reconnection via Listen()
// - Client successfully reconnects when daemon restarts
// - Encoder and decoder are recreated properly after reconnection
// - Connection is functional after reconnection (can subscribe)
// - Previous project subscription is restored after reconnection
func TestClient_ReconnectsAfterDaemonRestart(t *testing.T) {
	t.Parallel()

	// Setup controllable mock daemon
	socketPath, startFunc, stopFunc, messages := setupMockDaemonWithControl(t)

	// Start daemon and connect client
	err := startFunc()
	require.NoError(t, err)
	t.Logf("Initial daemon started")

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	// Set shorter retry parameters for faster test
	client.baseDelay = 100 * time.Millisecond
	client.maxRetries = 20
	client.setBatcherShutdownTimeoutForTest(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)

	// Drain initial subscribe message (project 0)
	select {
	case msg := <-messages:
		require.Equal(t, "subscribe", msg.Type)
		require.NotNil(t, msg.Subscribe)
		require.Equal(t, 0, msg.Subscribe.ProjectID)
		t.Logf("Client connected and subscribed initially")
	case <-time.After(2 * time.Second):
		require.Fail(t, "Timeout waiting for initial subscribe")
	}

	// Subscribe to a specific project before disconnect
	projectID := 5
	err = client.Subscribe(projectID)
	require.NoError(t, err)

	// Drain project subscribe message
	select {
	case msg := <-messages:
		require.Equal(t, "subscribe", msg.Type)
		require.NotNil(t, msg.Subscribe)
		require.Equal(t, projectID, msg.Subscribe.ProjectID)
		t.Logf("Subscribed to project %d", projectID)
	case <-time.After(2 * time.Second):
		require.Fail(t, "Timeout waiting for project subscribe")
	}

	// Verify connection state before stopping daemon
	client.mu.Lock()
	connBeforeStop := client.conn
	encoderBeforeStop := client.encoder
	decoderBeforeStop := client.decoder
	client.mu.Unlock()

	require.NotNil(t, connBeforeStop)
	require.NotNil(t, encoderBeforeStop)
	require.NotNil(t, decoderBeforeStop)
	t.Logf("Connection verified before daemon stop")

	// Start Listen() which will handle reconnection when daemon stops/restarts
	// Use the client's context so Listen will exit when we Close() the client
	eventChan, err := client.Listen(client.ctx)
	require.NoError(t, err)

	// Start draining events in background
	eventDone := make(chan struct{})
	go func() {
		defer close(eventDone)
		for range eventChan {
			// Drain events - we're just testing reconnection
		}
		t.Logf("Event channel closed")
	}()

	// Stop daemon to trigger disconnect, then restart it quickly in the background
	t.Logf("Stopping daemon to trigger disconnect...")
	stopFunc()

	// Restart daemon in a goroutine after a short delay
	// This simulates a daemon that restarts while the client is attempting reconnection
	go func() {
		time.Sleep(300 * time.Millisecond) // Wait for first reconnect attempt to be in progress
		if err := startFunc(); err != nil {
			assert.NoError(t, err, "Failed to restart daemon")
			return
		}
		t.Logf("Daemon restarted in background")
	}()

	// Wait for client to reconnect
	// The Listen loop's reconnect logic should detect the daemon is back
	// and successfully reconnect
	t.Logf("Waiting for client to reconnect...")

	// We should receive a new subscribe message after reconnection
	// The reconnect() function automatically subscribes to project 0,
	// then restores the previous subscription (project 5)
	reconnectTimeout := time.After(5 * time.Second)
	reconnectSuccess := false

	// Look for the restored project subscription as proof of reconnection
	for !reconnectSuccess {
		select {
		case msg := <-messages:
			t.Logf("Received message after restart: type=%s", msg.Type)
			if msg.Type == "subscribe" && msg.Subscribe != nil {
				t.Logf("  Subscribe message: ProjectID=%d", msg.Subscribe.ProjectID)
				// We're looking for the restored subscription to project 5
				if msg.Subscribe.ProjectID == projectID {
					reconnectSuccess = true
					t.Logf("Client reconnected and restored subscription to project %d", projectID)
				}
			}
		case <-reconnectTimeout:
			require.Fail(t, "Timeout waiting for client to reconnect and restore subscription")
		}
	}

	// Verify encoder/decoder were recreated
	client.mu.Lock()
	connAfterReconnect := client.conn
	encoderAfterReconnect := client.encoder
	decoderAfterReconnect := client.decoder
	client.mu.Unlock()

	assert.NotNil(t, connAfterReconnect)
	assert.NotNil(t, encoderAfterReconnect)
	assert.NotNil(t, decoderAfterReconnect)

	// Verify these are NEW instances (reconnect creates new connection/encoder/decoder)
	assert.NotEqual(t, connBeforeStop, connAfterReconnect)
	assert.NotEqual(t, encoderBeforeStop, encoderAfterReconnect)
	assert.NotEqual(t, decoderBeforeStop, decoderAfterReconnect)
	t.Logf("Encoder and decoder recreated after reconnection")

	// Verify connection is functional - try subscribing to a different project
	newProjectID := 7
	err = client.Subscribe(newProjectID)
	require.NoError(t, err)

	// Verify subscribe message is received
	select {
	case msg := <-messages:
		assert.Equal(t, "subscribe", msg.Type)
		require.NotNil(t, msg.Subscribe)
		assert.Equal(t, newProjectID, msg.Subscribe.ProjectID)
		t.Logf("Connection functional after reconnection - Subscribe works")
	case <-time.After(2 * time.Second):
		require.Fail(t, "Timeout waiting for subscribe after reconnection")
	}

	// Try sending an event to further verify functionality
	testEvent := Event{
		Type:      EventDatabaseChanged,
		ProjectID: newProjectID,
		Timestamp: time.Now(),
	}
	err = client.SendEvent(testEvent)
	require.NoError(t, err)

	// Wait for batching and verify event was sent
	time.Sleep(client.debounce + 100*time.Millisecond)
	select {
	case msg := <-messages:
		assert.Equal(t, "event", msg.Type)
		t.Logf("SendEvent works after reconnection")
	case <-time.After(2 * time.Second):
		require.Fail(t, "Timeout waiting for event after reconnection")
	}

	t.Logf("Test completed: client successfully reconnected after daemon restart")
}
