package events

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ============================================================================
// Test Helpers
// ============================================================================

// setupMockDaemon creates a simple mock daemon server for testing
func setupMockDaemon(t *testing.T) (string, net.Listener, chan Message) {
	t.Helper()

	// Create temp directory and socket path
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	// Create Unix socket listener
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create mock daemon listener: %v", err)
	}

	t.Cleanup(func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
	})

	// Channel to send messages received from client
	messages := make(chan Message, 10)

	// Accept connections in background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // Listener closed
			}

			go func(c net.Conn) {
				defer func() { _ = c.Close() }()
				decoder := json.NewDecoder(c)
				encoder := json.NewEncoder(c)

				for {
					var msg Message
					if err := decoder.Decode(&msg); err != nil {
						return
					}

					// Echo message to test channel
					select {
					case messages <- msg:
					default:
					}

					// Send ack for subscribe messages
					if msg.Type == "subscribe" {
						ackMsg := Message{
							Version: ProtocolVersion,
							Type:    "ack",
						}
						_ = encoder.Encode(ackMsg)
					}
				}
			}(conn)
		}
	}()

	return socketPath, listener, messages
}

// ============================================================================
// Client Creation Tests
// ============================================================================

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
	// Save original env var
	originalDebounce := os.Getenv("PASO_EVENT_DEBOUNCE_MS")
	defer func() { _ = os.Setenv("PASO_EVENT_DEBOUNCE_MS", originalDebounce) }()

	// Set custom debounce
	_ = os.Setenv("PASO_EVENT_DEBOUNCE_MS", "250")

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

// ============================================================================
// Connection Tests
// ============================================================================

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

// ============================================================================
// Subscribe Tests
// ============================================================================

func TestSubscribe_BeforeConnect(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Try to subscribe before connecting
	err = client.Subscribe(1)
	if err == nil {
		t.Error("Expected Subscribe to fail before connecting")
	}

	t.Logf("✓ Subscribe correctly fails before connection")
}

func TestSubscribe_AfterConnect(t *testing.T) {
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
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
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Subscribe to project 5
	if err := client.Subscribe(5); err != nil {
		t.Fatalf("Expected Subscribe to succeed, got error: %v", err)
	}

	// Wait for subscribe message
	select {
	case msg := <-messages:
		if msg.Type != "subscribe" {
			t.Errorf("Expected subscribe message, got: %s", msg.Type)
		}
		if msg.Subscribe == nil {
			t.Fatal("Expected subscribe message to have Subscribe field")
		}
		if msg.Subscribe.ProjectID != 5 {
			t.Errorf("Expected project ID 5, got %d", msg.Subscribe.ProjectID)
		}
		t.Logf("✓ Subscribe message sent correctly for project 5")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for subscribe message")
	}

	// Verify client's currentProjectID is updated
	client.mu.Lock()
	currentProject := client.currentProjectID
	client.mu.Unlock()

	if currentProject != 5 {
		t.Errorf("Expected currentProjectID to be 5, got %d", currentProject)
	}
}

func TestSubscribe_MultipleProjects(t *testing.T) {
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Drain initial subscribe message (project 0)
	select {
	case <-messages:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Subscribe to multiple projects
	projects := []int{1, 2, 3}
	for _, projectID := range projects {
		if err := client.Subscribe(projectID); err != nil {
			t.Fatalf("Failed to subscribe to project %d: %v", projectID, err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Verify we received all subscribe messages
	receivedProjects := make(map[int]bool)
	timeout := time.After(2 * time.Second)

	for i := 0; i < len(projects); i++ {
		select {
		case msg := <-messages:
			if msg.Type == "subscribe" && msg.Subscribe != nil {
				receivedProjects[msg.Subscribe.ProjectID] = true
			}
		case <-timeout:
			t.Fatal("Timeout waiting for subscribe messages")
		}
	}

	for _, projectID := range projects {
		if !receivedProjects[projectID] {
			t.Errorf("Did not receive subscribe message for project %d", projectID)
		}
	}

	t.Logf("✓ Multiple subscribe messages sent correctly")
}

// ============================================================================
// SendEvent Tests
// ============================================================================

func TestSendEvent_Success(t *testing.T) {
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Drain initial subscribe message
	select {
	case <-messages:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Send an event
	testEvent := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}

	if err := client.SendEvent(testEvent); err != nil {
		t.Fatalf("Expected SendEvent to succeed, got error: %v", err)
	}

	// Note: Events are batched, so we need to wait for debounce duration
	time.Sleep(client.debounce + 50*time.Millisecond)

	// Check if message was received (might be batched)
	select {
	case msg := <-messages:
		if msg.Type != "event" {
			t.Errorf("Expected event message, got: %s", msg.Type)
		}
		if msg.Event == nil {
			t.Fatal("Expected event message to have Event field")
		}
		t.Logf("✓ Event sent successfully: %+v", msg.Event)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event message")
	}
}

func TestSendEvent_BeforeConnect(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Send event before connecting - should succeed (queued)
	testEvent := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 1,
	}

	err = client.SendEvent(testEvent)
	if err != nil {
		t.Errorf("Expected SendEvent to succeed (queue event), got error: %v", err)
	}

	t.Logf("✓ SendEvent queues events before connection")
}

func TestSendEvent_Batching(t *testing.T) {
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	// Set short debounce for testing
	originalDebounce := os.Getenv("PASO_EVENT_DEBOUNCE_MS")
	_ = os.Setenv("PASO_EVENT_DEBOUNCE_MS", "50")
	defer func() { _ = os.Setenv("PASO_EVENT_DEBOUNCE_MS", originalDebounce) }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Drain initial subscribe message
	select {
	case <-messages:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Send multiple events rapidly
	numEvents := 5
	for i := 0; i < numEvents; i++ {
		testEvent := Event{
			Type:      EventDatabaseChanged,
			ProjectID: i,
		}
		if err := client.SendEvent(testEvent); err != nil {
			t.Fatalf("Failed to send event %d: %v", i, err)
		}
	}

	// Wait for batch to be sent
	time.Sleep(client.debounce + 100*time.Millisecond)

	// Should receive at least one message (events might be batched)
	select {
	case msg := <-messages:
		if msg.Type != "event" {
			t.Errorf("Expected event message, got: %s", msg.Type)
		}
		t.Logf("✓ Batched events sent successfully")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for batched events")
	}
}

// ============================================================================
// Close Tests
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

// ============================================================================
// Write Deadline Tests (Bug Fix Regression Tests)
// ============================================================================

// TestSubscribe_AfterLongDelay tests that Subscribe works even after the write
// deadline has expired from a previous operation. This is the core bug that was
// fixed: lingering write deadlines caused Subscribe to fail with "i/o timeout"
// when called after the deadline expired (e.g., when navigating back to previous projects).
//
// Bug context: When switching between projects in the TUI, especially after delays
// >5 seconds, the Subscribe() call would immediately fail because sendToSocket()
// set a deadline but never cleared it, and Subscribe() didn't set its own deadline.
func TestSubscribe_AfterLongDelay(t *testing.T) {
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Set short write deadline for faster test execution
	client.setWriteDeadlineForTest(500 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Drain initial subscribe message (project 0)
	select {
	case msg := <-messages:
		if msg.Type != "subscribe" {
			t.Fatalf("Expected initial subscribe, got: %s", msg.Type)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Send an event to trigger setting a write deadline
	testEvent := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}
	if err := client.SendEvent(testEvent); err != nil {
		t.Fatalf("Failed to send event: %v", err)
	}

	// Wait for event to be sent (batching)
	time.Sleep(client.debounce + 100*time.Millisecond)

	// Drain the event message
	select {
	case msg := <-messages:
		if msg.Type != "event" {
			t.Fatalf("Expected event message, got: %s", msg.Type)
		}
		t.Logf("✓ Event sent successfully")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event")
	}

	// Wait LONGER than the write deadline (600ms > 500ms)
	// This simulates user navigating between projects with a delay
	t.Logf("Waiting for write deadline to expire...")
	time.Sleep(600 * time.Millisecond)

	// NOW try to subscribe - this is where the bug would occur
	// Without the fix, this would fail with "i/o timeout" because:
	// 1. sendToSocket() set a 500ms deadline
	// 2. We waited 600ms (deadline expired)
	// 3. Subscribe() would inherit the expired deadline and immediately fail
	t.Logf("Attempting Subscribe after deadline expired...")
	if err := client.Subscribe(2); err != nil {
		t.Fatalf("Subscribe failed after deadline expired: %v (BUG REPRODUCED)", err)
	}

	// Verify subscribe message was sent
	select {
	case msg := <-messages:
		if msg.Type != "subscribe" {
			t.Errorf("Expected subscribe message, got: %s", msg.Type)
		}
		if msg.Subscribe == nil || msg.Subscribe.ProjectID != 2 {
			t.Errorf("Expected subscribe to project 2, got: %+v", msg.Subscribe)
		}
		t.Logf("✓ Subscribe succeeded after deadline expired (BUG FIXED)")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for subscribe message")
	}

	// Verify we can still send events after this
	testEvent2 := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 2,
		Timestamp: time.Now(),
	}
	if err := client.SendEvent(testEvent2); err != nil {
		t.Fatalf("Failed to send event after subscribe: %v", err)
	}

	time.Sleep(client.debounce + 100*time.Millisecond)

	select {
	case msg := <-messages:
		if msg.Type != "event" {
			t.Errorf("Expected event message, got: %s", msg.Type)
		}
		t.Logf("✓ Events still work after deadline recovery")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event after subscribe")
	}
}

// TestSubscribe_ClearsWriteDeadline verifies that Subscribe properly clears the
// write deadline after encoding, so it doesn't affect future operations.
func TestSubscribe_ClearsWriteDeadline(t *testing.T) {
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Set short write deadline for faster test
	client.setWriteDeadlineForTest(500 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Drain initial subscribe (project 0)
	select {
	case <-messages:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Subscribe to project 1
	if err := client.Subscribe(1); err != nil {
		t.Fatalf("First subscribe failed: %v", err)
	}

	// Drain subscribe message
	select {
	case msg := <-messages:
		if msg.Type != "subscribe" || msg.Subscribe.ProjectID != 1 {
			t.Fatalf("Expected subscribe to project 1, got: %+v", msg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for first subscribe")
	}

	// Wait longer than deadline
	time.Sleep(600 * time.Millisecond)

	// Subscribe to project 2 - should work because deadline was cleared
	if err := client.Subscribe(2); err != nil {
		t.Fatalf("Second subscribe failed after deadline: %v", err)
	}

	select {
	case msg := <-messages:
		if msg.Type != "subscribe" || msg.Subscribe.ProjectID != 2 {
			t.Fatalf("Expected subscribe to project 2, got: %+v", msg)
		}
		t.Logf("✓ Subscribe clears write deadline correctly")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for second subscribe")
	}
}

// TestSubscribe_MultipleRapidChanges tests rapidly switching between projects
// to ensure no race conditions or deadline issues occur.
func TestSubscribe_MultipleRapidChanges(t *testing.T) {
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Drain initial subscribe (project 0)
	select {
	case <-messages:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Rapidly switch between 5 projects
	projects := []int{1, 2, 3, 4, 5, 4, 3, 2, 1}
	for _, projectID := range projects {
		if err := client.Subscribe(projectID); err != nil {
			t.Fatalf("Subscribe to project %d failed: %v", projectID, err)
		}
		time.Sleep(10 * time.Millisecond) // Very short delay
	}

	// Verify we received all subscribe messages
	receivedCount := 0
	timeout := time.After(2 * time.Second)

	for receivedCount < len(projects) {
		select {
		case msg := <-messages:
			if msg.Type == "subscribe" {
				receivedCount++
			}
		case <-timeout:
			t.Fatalf("Only received %d/%d subscribe messages", receivedCount, len(projects))
		}
	}

	t.Logf("✓ All %d rapid subscribe changes succeeded", len(projects))
}

// TestSendEvent_ClearsDeadline verifies that SendEvent (via sendToSocket) properly
// clears the write deadline after encoding.
func TestSendEvent_ClearsDeadline(t *testing.T) {
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Set short write deadline for faster test
	client.setWriteDeadlineForTest(500 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Drain initial subscribe (project 0)
	select {
	case <-messages:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Send first event
	event1 := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}
	if err := client.SendEvent(event1); err != nil {
		t.Fatalf("First SendEvent failed: %v", err)
	}

	// Wait for batching
	time.Sleep(client.debounce + 100*time.Millisecond)

	// Drain first event
	select {
	case msg := <-messages:
		if msg.Type != "event" {
			t.Fatalf("Expected event, got: %s", msg.Type)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for first event")
	}

	// Wait longer than deadline
	time.Sleep(600 * time.Millisecond)

	// Send second event - should work because deadline was cleared
	event2 := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 2,
		Timestamp: time.Now(),
	}
	if err := client.SendEvent(event2); err != nil {
		t.Fatalf("Second SendEvent failed after deadline: %v", err)
	}

	time.Sleep(client.debounce + 100*time.Millisecond)

	select {
	case msg := <-messages:
		if msg.Type != "event" {
			t.Fatalf("Expected event, got: %s", msg.Type)
		}
		t.Logf("✓ SendEvent clears write deadline correctly")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for second event")
	}
}
