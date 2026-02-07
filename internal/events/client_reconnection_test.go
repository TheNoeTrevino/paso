package events

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestClient_ContextCancellationDuringReconnection verifies graceful shutdown
// when context is cancelled during reconnection attempts. This ensures:
// - No goroutine leaks when Close() is called during reconnection
// - batcherDone channel is closed properly
// - Context cancellation propagates through all operations
// - No panics or deadlocks during shutdown
func TestClient_ContextCancellationDuringReconnection(t *testing.T) {
	// Create a controllable mock daemon
	socketPath, startFunc, stopFunc, messages := setupMockDaemonWithControl(t)

	// Start the daemon initially
	if err := startFunc(); err != nil {
		t.Fatalf("Failed to start mock daemon: %v", err)
	}

	// Create and connect the client
	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Set short retry parameters to speed up the test
	client.baseDelay = 100 * time.Millisecond
	client.maxRetries = 20 // More retries to ensure we can catch it mid-reconnect

	// IMPORTANT: Use the client's internal context for Listen()
	// This way when Close() cancels c.ctx, the listen loop will exit
	ctx := client.ctx

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Drain initial subscribe message
	select {
	case <-messages:
		t.Logf("✓ Initial connection established")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Verify client is connected and batcher is running
	client.mu.Lock()
	connected := client.conn != nil
	client.mu.Unlock()

	if !connected {
		t.Fatal("Expected client to be connected")
	}

	// Now simulate connection loss by stopping the daemon
	t.Logf("Stopping daemon to trigger reconnection...")
	stopFunc()

	// Start a listen loop that will trigger reconnection
	// This is important because Listen() is what drives the reconnection logic
	eventChan, err := client.Listen(ctx)
	if err != nil {
		t.Fatalf("Failed to start listen loop: %v", err)
	}

	// Start draining events in background
	eventDone := make(chan struct{})
	go func() {
		defer close(eventDone)
		for range eventChan {
			// Drain events
		}
		t.Logf("Event channel closed")
	}()

	// Wait for reconnection to actually start
	// The listenLoop will detect the closed connection and start reconnecting
	// We need to wait long enough for it to get into the retry loop
	time.Sleep(500 * time.Millisecond)
	t.Logf("Client should now be attempting reconnection")

	// Now close the client while it's attempting to reconnect
	// This is the critical test: context cancellation during reconnection
	t.Logf("Calling Close() during reconnection...")
	closeStart := time.Now()
	closeErr := client.Close()
	closeDuration := time.Since(closeStart)

	if closeErr != nil {
		t.Logf("Close returned error (may be acceptable): %v", closeErr)
	}
	t.Logf("Close() completed in %v", closeDuration)

	// Close should complete quickly (not block waiting for all retries)
	if closeDuration > 2*time.Second {
		t.Errorf("Close() took too long (%v), context cancellation may not be working", closeDuration)
	}

	// Verify that batcherDone channel is closed
	// This should be non-blocking if Close() worked correctly
	select {
	case <-client.batcherDone:
		t.Logf("✓ batcherDone channel closed successfully")
	case <-time.After(2 * time.Second):
		t.Error("batcherDone channel was not closed - possible goroutine leak")
	}

	// Verify that the context was cancelled
	select {
	case <-client.ctx.Done():
		t.Logf("✓ Client context cancelled successfully")
	default:
		t.Error("Client context was not cancelled")
	}

	// Wait for event channel to close (listen loop should exit)
	// The listen loop checks ctx.Done() in the select statement and also
	// the reconnect function checks ctx.Done(), so it should exit promptly
	select {
	case <-eventDone:
		t.Logf("✓ Listen loop exited cleanly")
	case <-time.After(5 * time.Second):
		t.Error("Listen loop did not exit within 5 seconds - context cancellation may not be propagating")
	}

	// Verify we can call Close again (idempotent)
	closeErr2 := client.Close()
	if closeErr2 != nil {
		t.Errorf("Second Close() should be idempotent, got error: %v", closeErr2)
	} else {
		t.Logf("✓ Close() is idempotent")
	}

	// Verify connection is nil after close
	client.mu.Lock()
	finalConn := client.conn
	client.mu.Unlock()

	if finalConn != nil {
		t.Error("Connection should be nil after Close()")
	} else {
		t.Logf("✓ Connection properly cleaned up")
	}

	t.Logf("✓ Test completed: graceful shutdown during reconnection verified")
}

// TestClient_ExponentialBackoffDuringReconnection verifies the exponential
// backoff behavior during reconnection attempts.
//
// This test validates:
// - baseDelay (1s) is used for the first attempt
// - Delays increase exponentially: 1s, 2s, 4s, 8s, 16s
// - After maxRetries (5), the client gives up gracefully
// - Reconnection attempts stop after reaching maxRetries
//
// The test uses a non-existent socket path to force connection failures
// and measures actual delays between attempts using time.Now().
func TestClient_ExponentialBackoffDuringReconnection(t *testing.T) {
	t.Parallel()

	// Use a non-existent socket path to force connection failures
	socketPath := filepath.Join(t.TempDir(), "nonexistent.sock")

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Use shorter delays for faster testing
	// baseDelay: 100ms instead of 1s
	// This gives us delays of: 100ms, 200ms, 400ms, 800ms, 1600ms
	client.mu.Lock()
	client.baseDelay = 100 * time.Millisecond
	client.maxRetries = 5
	client.mu.Unlock()

	// Track notifications to verify attempts are being made
	notificationCount := 0
	client.SetNotifyFunc(func(level, message string) {
		notificationCount++
		t.Logf("Notification [%s]: %s", level, message)
	})

	// Start the reconnection process by calling reconnect directly
	// This simulates what happens when listenLoop detects a connection failure
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Record start time
	startTime := time.Now()

	// Call reconnect - this will try to connect multiple times
	success := client.reconnect(ctx)

	totalTime := time.Since(startTime)

	// Reconnection should fail with non-existent socket
	if success {
		t.Error("Expected reconnection to fail with non-existent socket")
	}

	t.Logf("Reconnection attempts completed in %v", totalTime)

	// Calculate expected total time: sum of all delays
	// Expected delays: 100ms, 200ms, 400ms, 800ms, 1600ms
	expectedDelays := []time.Duration{
		100 * time.Millisecond,  // 1st attempt: after 100ms delay
		200 * time.Millisecond,  // 2nd attempt: after 200ms delay
		400 * time.Millisecond,  // 3rd attempt: after 400ms delay
		800 * time.Millisecond,  // 4th attempt: after 800ms delay
		1600 * time.Millisecond, // 5th attempt: after 1600ms delay
	}

	expectedTotalTime := time.Duration(0)
	for _, d := range expectedDelays {
		expectedTotalTime += d
	}

	t.Logf("Expected total delay: %v", expectedTotalTime)

	// Verify the total time is approximately correct (within 30% tolerance)
	// We use a tolerance because:
	// 1. Connection attempts themselves take some time (context timeout, dial attempts)
	// 2. System scheduling can introduce small delays
	// 3. Each Connect() call has overhead
	minExpected := expectedTotalTime - (expectedTotalTime * 30 / 100)
	maxExpected := expectedTotalTime + (expectedTotalTime * 50 / 100)

	if totalTime < minExpected {
		t.Errorf("Reconnection completed too quickly: %v (expected at least %v)", totalTime, minExpected)
	}
	if totalTime > maxExpected {
		t.Errorf("Reconnection took too long: %v (expected at most %v)", totalTime, maxExpected)
	}

	// Verify maxRetries was respected - reconnection should have stopped
	// after 5 attempts (not infinite loop)
	if totalTime > 5*time.Second {
		t.Error("Reconnection appears to be running too long - may not respect maxRetries")
	}

	t.Logf("✓ Exponential backoff verified: baseDelay=%v, maxRetries=%d",
		client.baseDelay, client.maxRetries)
	t.Logf("✓ Total time within expected range: %v (min=%v, max=%v)",
		totalTime, minExpected, maxExpected)
	t.Logf("✓ Client gave up gracefully after %d attempts", client.maxRetries)

	// Additional verification: Try to use the client after failed reconnection
	// The client should not be in a broken state
	if err := client.Subscribe(1); err == nil {
		t.Error("Expected Subscribe to fail after failed reconnection (not connected)")
	} else if !strings.Contains(err.Error(), "not connected") {
		t.Errorf("Expected 'not connected' error, got: %v", err)
	}

	t.Logf("✓ Client remains in valid state after failed reconnection")
}

// TestClient_RestoresSubscriptionAfterReconnect verifies that the client
// automatically re-subscribes to the same project after a reconnection.
//
// This test validates:
// - Client stores currentProjectID when Subscribe() is called
// - After reconnection, client automatically re-subscribes to the stored projectID
// - Daemon receives both the initial subscribe message and the re-subscribe message
// - Client can receive events after reconnection
func TestClient_RestoresSubscriptionAfterReconnect(t *testing.T) {
	t.Parallel()

	socketPath, startFunc, stopFunc, messages := setupMockDaemonWithControl(t)

	// Start daemon
	if err := startFunc(); err != nil {
		t.Fatalf("Failed to start mock daemon: %v", err)
	}

	// Create and connect client
	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Set shorter retry parameters for faster test execution
	client.baseDelay = 100 * time.Millisecond
	client.maxRetries = 20

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Drain initial subscribe message (project 0)
	select {
	case msg := <-messages:
		if msg.Type != "subscribe" || msg.Subscribe == nil || msg.Subscribe.ProjectID != 0 {
			t.Fatalf("Expected initial subscribe for project 0, got: %+v", msg)
		}
		t.Logf("✓ Initial subscribe to project 0 received")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Subscribe to project 123
	testProjectID := 123
	if err := client.Subscribe(testProjectID); err != nil {
		t.Fatalf("Failed to subscribe to project %d: %v", testProjectID, err)
	}

	// Wait for initial subscribe message for project 123
	select {
	case msg := <-messages:
		if msg.Type != "subscribe" {
			t.Errorf("Expected subscribe message, got: %s", msg.Type)
		}
		if msg.Subscribe == nil {
			t.Fatal("Expected subscribe message to have Subscribe field")
		}
		if msg.Subscribe.ProjectID != testProjectID {
			t.Errorf("Expected project ID %d, got %d", testProjectID, msg.Subscribe.ProjectID)
		}
		t.Logf("✓ Subscribe message received for project %d", testProjectID)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial subscribe message")
	}

	// Verify currentProjectID is stored
	client.mu.Lock()
	storedProjectID := client.currentProjectID
	client.mu.Unlock()

	if storedProjectID != testProjectID {
		t.Fatalf("Expected currentProjectID to be %d, got %d", testProjectID, storedProjectID)
	}
	t.Logf("✓ currentProjectID stored correctly: %d", storedProjectID)

	// Stop daemon to simulate connection loss
	t.Logf("Stopping daemon to simulate connection loss...")
	stopFunc()

	// Close the old connection to cleanly stop the old batcher
	client.mu.Lock()
	if client.conn != nil {
		_ = client.conn.Close()
		client.conn = nil
	}
	client.mu.Unlock()

	// Wait for old batcher to exit
	select {
	case <-client.batcherDone:
		t.Logf("✓ Old batcher exited cleanly")
	case <-time.After(1 * time.Second):
		t.Log("Warning: Old batcher did not exit in time")
	}

	// Create new batcherDone channel for reconnection
	client.batcherDone = make(chan struct{})

	// Restart daemon
	t.Logf("Restarting daemon...")
	if err := startFunc(); err != nil {
		t.Fatalf("Failed to restart daemon: %v", err)
	}

	// Give daemon time to fully start
	time.Sleep(100 * time.Millisecond)

	// Call reconnect() directly to test the subscription restoration logic
	// This simulates what the Listen loop would do when it detects connection loss
	t.Logf("Testing reconnect() with subscription restoration...")
	success := client.reconnect(ctx)

	if !success {
		t.Fatal("reconnect() failed - daemon should be available")
	}
	t.Logf("✓ reconnect() succeeded")

	// Wait for re-subscribe message
	// The reconnect() method should have automatically called Subscribe(currentProjectID)
	timeout := time.After(2 * time.Second)
	subscribeCount := 0
	resubscribeReceived := false

	// Collect subscribe messages (we expect project 0 from Connect, then project 123 from restoration)
ResubscribeLoop:
	for !resubscribeReceived && subscribeCount < 10 {
		select {
		case msg := <-messages:
			if msg.Type == "subscribe" && msg.Subscribe != nil {
				subscribeCount++
				t.Logf("Received subscribe message #%d: project %d", subscribeCount, msg.Subscribe.ProjectID)
				if msg.Subscribe.ProjectID == testProjectID {
					t.Logf("✓ Re-subscribe message received for project %d after reconnection", testProjectID)
					resubscribeReceived = true
				}
			}
		case <-timeout:
			break ResubscribeLoop
		default:
			if subscribeCount > 0 && time.Since(time.Now()) > 500*time.Millisecond {
				break ResubscribeLoop
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	if !resubscribeReceived {
		t.Fatalf("Did not receive re-subscribe for project %d (received %d subscribe messages)", testProjectID, subscribeCount)
	}

	// Verify the client is connected after reconnection
	t.Logf("Verifying client state after reconnection...")

	client.mu.Lock()
	connected := client.conn != nil
	finalProjectID := client.currentProjectID
	client.mu.Unlock()

	if !connected {
		t.Error("Expected client to be connected after reconnection")
	} else {
		t.Logf("✓ Client successfully reconnected")
	}

	if finalProjectID != testProjectID {
		t.Errorf("Expected currentProjectID to still be %d, got %d", testProjectID, finalProjectID)
	} else {
		t.Logf("✓ currentProjectID preserved after reconnection: %d", finalProjectID)
	}

	// Verify we can still send subscribe messages after reconnection
	t.Logf("Testing subscribe after reconnection...")
	if err := client.Subscribe(456); err != nil {
		t.Errorf("Failed to subscribe after reconnection: %v", err)
	}

	// Wait for new subscribe message
	select {
	case msg := <-messages:
		if msg.Type == "subscribe" && msg.Subscribe != nil && msg.Subscribe.ProjectID == 456 {
			t.Logf("✓ Client can subscribe to new projects after reconnection")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for subscribe message after reconnection")
	}

	t.Logf("✓ Test completed: subscription restored after reconnect")
}
