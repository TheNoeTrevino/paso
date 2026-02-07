package events

import (
	"context"
	"os"
	"testing"
	"time"
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
	if err := startFunc(); err != nil {
		t.Fatalf("Failed to start mock daemon: %v", err)
	}
	t.Logf("✓ Mock daemon started")

	// Create client
	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Set very short retry parameters for faster testing
	client.baseDelay = 50 * time.Millisecond
	client.maxRetries = 2
	client.setBatcherShutdownTimeoutForTest(25 * time.Millisecond)

	// Connect
	connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Connect(connectCtx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	t.Logf("✓ Client connected")

	// Drain initial subscribe
	select {
	case <-messages:
		t.Logf("✓ Initial subscribe received")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for subscribe")
	}

	// Start Listen with client's internal context
	eventChan, err := client.Listen(client.ctx)
	if err != nil {
		t.Fatalf("Failed to start listen: %v", err)
	}

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
		t.Fatalf("Failed to remove socket: %v", err)
	}

	// Wait for Listen loop to detect disconnect and exit
	// With baseDelay=50ms and maxRetries=2: ~150ms for retries
	// Plus time to detect EOF (should be immediate when decoder reads)
	select {
	case <-done:
		elapsed := time.Since(start)
		t.Logf("✓ Client detected disconnect in %v", elapsed)

		// Generous timeout: should complete in < 1 second
		if elapsed > 1*time.Second {
			t.Logf("⚠ Detection took longer than expected")
		}

	case <-time.After(10 * time.Second):
		t.Fatal("Client did not detect disconnect within 10s")
	}

	t.Logf("✓ Test completed: disconnect detection verified")
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
	if err := startFunc(); err != nil {
		t.Fatalf("Failed to start initial daemon: %v", err)
	}
	t.Logf("✓ Initial daemon started")

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Set shorter retry parameters for faster test
	client.baseDelay = 100 * time.Millisecond
	client.maxRetries = 20
	client.setBatcherShutdownTimeoutForTest(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to initial connect: %v", err)
	}

	// Drain initial subscribe message (project 0)
	select {
	case msg := <-messages:
		if msg.Type != "subscribe" || msg.Subscribe == nil || msg.Subscribe.ProjectID != 0 {
			t.Fatalf("Expected initial subscribe for project 0, got: %+v", msg)
		}
		t.Logf("✓ Client connected and subscribed initially")
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Subscribe to a specific project before disconnect
	projectID := 5
	if err := client.Subscribe(projectID); err != nil {
		t.Fatalf("Failed to subscribe to project %d: %v", projectID, err)
	}

	// Drain project subscribe message
	select {
	case msg := <-messages:
		if msg.Type != "subscribe" || msg.Subscribe == nil || msg.Subscribe.ProjectID != projectID {
			t.Fatalf("Expected subscribe for project %d, got: %+v", projectID, msg)
		}
		t.Logf("✓ Subscribed to project %d", projectID)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for project subscribe")
	}

	// Verify connection state before stopping daemon
	client.mu.Lock()
	connBeforeStop := client.conn
	encoderBeforeStop := client.encoder
	decoderBeforeStop := client.decoder
	client.mu.Unlock()

	if connBeforeStop == nil {
		t.Fatal("Expected connection to be established before stop")
	}
	if encoderBeforeStop == nil || decoderBeforeStop == nil {
		t.Fatal("Expected encoder/decoder to be set before stop")
	}
	t.Logf("✓ Connection verified before daemon stop")

	// Start Listen() which will handle reconnection when daemon stops/restarts
	// Use the client's context so Listen will exit when we Close() the client
	eventChan, err := client.Listen(client.ctx)
	if err != nil {
		t.Fatalf("Failed to start Listen: %v", err)
	}

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
			t.Errorf("Failed to restart daemon: %v", err)
			return
		}
		t.Logf("✓ Daemon restarted in background")
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
					t.Logf("✓ Client reconnected and restored subscription to project %d", projectID)
				}
			}
		case <-reconnectTimeout:
			t.Fatal("Timeout waiting for client to reconnect and restore subscription")
		}
	}

	// Verify encoder/decoder were recreated
	client.mu.Lock()
	connAfterReconnect := client.conn
	encoderAfterReconnect := client.encoder
	decoderAfterReconnect := client.decoder
	client.mu.Unlock()

	if connAfterReconnect == nil {
		t.Error("Expected connection to be re-established after reconnect")
	}
	if encoderAfterReconnect == nil || decoderAfterReconnect == nil {
		t.Error("Expected encoder/decoder to be recreated after reconnect")
	}

	// Verify these are NEW instances (reconnect creates new connection/encoder/decoder)
	if connAfterReconnect == connBeforeStop {
		t.Error("Expected new connection instance after reconnect")
	}
	if encoderAfterReconnect == encoderBeforeStop {
		t.Error("Expected new encoder instance after reconnect")
	}
	if decoderAfterReconnect == decoderBeforeStop {
		t.Error("Expected new decoder instance after reconnect")
	}
	t.Logf("✓ Encoder and decoder recreated after reconnection")

	// Verify connection is functional - try subscribing to a different project
	newProjectID := 7
	if err := client.Subscribe(newProjectID); err != nil {
		t.Fatalf("Subscribe failed after reconnection: %v", err)
	}

	// Verify subscribe message is received
	select {
	case msg := <-messages:
		if msg.Type != "subscribe" {
			t.Errorf("Expected subscribe message after reconnect, got: %s", msg.Type)
		}
		if msg.Subscribe == nil || msg.Subscribe.ProjectID != newProjectID {
			t.Errorf("Expected subscribe to project %d, got: %+v", newProjectID, msg.Subscribe)
		}
		t.Logf("✓ Connection functional after reconnection - Subscribe works")
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for subscribe after reconnection")
	}

	// Try sending an event to further verify functionality
	testEvent := Event{
		Type:      EventDatabaseChanged,
		ProjectID: newProjectID,
		Timestamp: time.Now(),
	}
	if err := client.SendEvent(testEvent); err != nil {
		t.Fatalf("SendEvent failed after reconnection: %v", err)
	}

	// Wait for batching and verify event was sent
	time.Sleep(client.debounce + 100*time.Millisecond)
	select {
	case msg := <-messages:
		if msg.Type != "event" {
			t.Errorf("Expected event message, got: %s", msg.Type)
		}
		t.Logf("✓ SendEvent works after reconnection")
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event after reconnection")
	}

	t.Logf("✓ Test completed: client successfully reconnected after daemon restart")
}
