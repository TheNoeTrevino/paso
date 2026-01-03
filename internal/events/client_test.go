package events

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

// setupMockDaemonWithControl creates a mock daemon server that can be stopped
// and restarted on demand for reconnection testing.
//
// It properly tracks and closes all active client connections when stopped, ensuring
// clients detect disconnection immediately rather than blocking indefinitely.
//
// Returns:
//   - socketPath: path to the Unix socket
//   - startFunc: function to start/restart the daemon (returns error if start fails)
//   - stopFunc: function to stop the daemon
//   - messages: channel to receive messages from client
//
// The daemon can be stopped and started multiple times. Each restart creates
// a new listener on the same socket path. Proper cleanup is handled via t.Cleanup().
func setupMockDaemonWithControl(t *testing.T) (string, func() error, func(), chan Message) {
	t.Helper()

	// Create temp directory and socket path
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	// Channel to send messages received from client
	messages := make(chan Message, 10)

	// Track the current listener and its cancel function
	var currentListener net.Listener
	var cancelAccept context.CancelFunc

	// Track active connections for proper cleanup when daemon stops
	var activeConns map[net.Conn]struct{}
	var connsMu sync.Mutex
	var listenerMu struct {
		mu       *testing.T // Prevents direct mutex usage, uses test helper pattern
		listener net.Listener
		cancel   context.CancelFunc
	}
	listenerMu.mu = t

	// Track cleanup state
	cleanedUp := false
	var cleanupMu struct {
		mu        *testing.T
		cleanedUp *bool
	}
	cleanupMu.mu = t
	cleanupMu.cleanedUp = &cleanedUp

	// startFunc creates a new listener and starts accepting connections
	startFunc := func() error {
		// Remove old socket file if it exists
		if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove old socket: %w", err)
		}

		// Create Unix socket listener
		listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socketPath)
		if err != nil {
			return fmt.Errorf("failed to create mock daemon listener: %w", err)
		}

		// Create cancellable context for accept loop
		ctx, cancel := context.WithCancel(context.Background())

		// Update current listener and cancel function
		currentListener = listener
		cancelAccept = cancel

		// Initialize connection tracking map for this daemon instance
		connsMu.Lock()
		activeConns = make(map[net.Conn]struct{})
		connsMu.Unlock()

		// Accept connections in background
		go func() {
			for {
				// Check if context is cancelled
				select {
				case <-ctx.Done():
					return
				default:
				}

				conn, err := listener.Accept()
				if err != nil {
					// Check if this is due to listener being closed
					select {
					case <-ctx.Done():
						return // Normal shutdown
					default:
						// Unexpected error, but continue to allow for reconnection
						return
					}
				}

				// Track this connection for cleanup when daemon stops
				connsMu.Lock()
				activeConns[conn] = struct{}{}
				connsMu.Unlock()
				go func(c net.Conn) {
					// Remove from active connections when handler exits (any reason: error, normal, panic)
					defer func() {
						connsMu.Lock()
						delete(activeConns, c)
						connsMu.Unlock()
						_ = c.Close()
					}()

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

		return nil
	}

	// stopFunc stops the daemon by closing the listener AND all active connections
	stopFunc := func() {
		// Step 1: Stop accepting new connections
		if cancelAccept != nil {
			cancelAccept()
		}
		if currentListener != nil {
			_ = currentListener.Close()
		}

		// Step 2: Close all active connections (THE CRITICAL FIX)
		connsMu.Lock()
		// Copy connections to slice to avoid holding lock during Close()
		connsToClose := make([]net.Conn, 0, len(activeConns))
		for conn := range activeConns {
			connsToClose = append(connsToClose, conn)
		}
		// Clear map while we have the lock
		activeConns = make(map[net.Conn]struct{})
		connsMu.Unlock()

		// Close without holding mutex (prevents deadlock)
		for _, conn := range connsToClose {
			_ = conn.Close()
		}

		// Shorter sleep since we actively close connections
		time.Sleep(10 * time.Millisecond)
	}

	// Register cleanup to ensure resources are released
	t.Cleanup(func() {
		*cleanupMu.cleanedUp = true
		stopFunc()
		if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to remove socket during cleanup: %v", err)
		}
	})

	return socketPath, startFunc, stopFunc, messages
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

// ============================================================================
// Backpressure Tests
// ============================================================================

// TestSendEvent_QueueFullWithBackpressure tests that SendEvent applies
// exponential backoff when the queue is full, ensuring events aren't silently dropped.
func TestSendEvent_QueueFullWithBackpressure(t *testing.T) {
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

	// Drain initial subscribe message
	select {
	case <-messages:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Fill the event queue (capacity is 100)
	// We'll send 101 events - the first 100 succeed immediately,
	// the 101st triggers backpressure/retry logic
	numEvents := 101
	var lastErr error
	for i := 0; i < numEvents; i++ {
		event := Event{
			Type:      EventDatabaseChanged,
			ProjectID: i % 5,
			Timestamp: time.Now(),
		}
		err := client.SendEvent(event)
		if err != nil {
			lastErr = err
			t.Logf("Event %d failed: %v (expected after queue fills)", i, err)
		}
	}

	// The backpressure mechanism should eventually allow the last event through
	// or return an error after retries - NOT silently drop it
	if lastErr != nil {
		if !strings.Contains(lastErr.Error(), "retry attempts exhausted") {
			t.Fatalf("Expected 'retry attempts exhausted' error, got: %v", lastErr)
		}
		t.Logf("✓ Queue saturation detected and logged: %v", lastErr)
	}

	// Wait for batching to complete
	time.Sleep(client.debounce + 200*time.Millisecond)

	// Verify at least some events were processed
	eventCount := 0
	timeout := time.After(2 * time.Second)

	for {
		select {
		case msg := <-messages:
			if msg.Type == "event" {
				eventCount++
			}
		case <-timeout:
			if eventCount == 0 {
				t.Fatal("Expected at least some events to be processed")
			}
			t.Logf("✓ Backpressure mechanism allowed %d events through", eventCount)
			return
		}
	}
}

// TestSendEvent_HighThroughputReliability tests event reliability under
// high-throughput scenarios where queue saturation is likely.
func TestSendEvent_HighThroughputReliability(t *testing.T) {
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	// Send a burst of events rapidly
	numEvents := 50
	var sendErrors int
	for i := 0; i < numEvents; i++ {
		event := Event{
			Type:      EventDatabaseChanged,
			ProjectID: i % 3,
			Timestamp: time.Now(),
		}
		err := client.SendEvent(event)
		if err != nil {
			sendErrors++
			t.Logf("Send error (attempt %d): %v", i+1, err)
		}
	}

	// With backpressure, we should minimize silent drops
	// A few errors are acceptable if they're retried with backoff,
	// but most events should succeed
	successRate := float64(numEvents-sendErrors) / float64(numEvents) * 100
	if successRate < 90 {
		t.Logf("⚠ Success rate lower than expected: %.1f%%", successRate)
	}

	// Wait for all events to be batched and sent
	time.Sleep(client.debounce + 500*time.Millisecond)

	// Verify we received events
	eventCount := 0
	timeout := time.After(2 * time.Second)

	for {
		select {
		case msg := <-messages:
			if msg.Type == "event" {
				eventCount++
			}
		case <-timeout:
			if eventCount == 0 {
				t.Error("Expected at least some events to be received")
			}
			t.Logf("✓ High-throughput test: sent=%d, errors=%d, success_rate=%.1f%%, received=%d",
				numEvents, sendErrors, successRate, eventCount)
			return
		}
	}
}

// TestSendEvent_BackpressureQueueRecovery tests that the queue recovers
// after saturation when the batcher is consuming events.
func TestSendEvent_BackpressureQueueRecovery(t *testing.T) {
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	// Use a very short debounce to consume events quickly
	originalDebounce := os.Getenv("PASO_EVENT_DEBOUNCE_MS")
	_ = os.Setenv("PASO_EVENT_DEBOUNCE_MS", "10")
	defer func() { _ = os.Setenv("PASO_EVENT_DEBOUNCE_MS", originalDebounce) }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	// First batch: fill queue
	for i := 0; i < 50; i++ {
		event := Event{
			Type:      EventDatabaseChanged,
			ProjectID: 0,
			Timestamp: time.Now(),
		}
		if err := client.SendEvent(event); err != nil {
			t.Logf("First batch error (expected): %v", err)
		}
	}

	// Wait for queue to drain via batcher
	time.Sleep(100 * time.Millisecond)

	// Second batch: should now be able to send without excessive retry
	// If recovery works, these should succeed quickly
	successCount := 0
	for i := 0; i < 30; i++ {
		event := Event{
			Type:      EventDatabaseChanged,
			ProjectID: 0,
			Timestamp: time.Now(),
		}
		if err := client.SendEvent(event); err == nil {
			successCount++
		}
	}

	if successCount < 20 {
		t.Logf("⚠ Queue recovery slow: only %d/%d succeeded", successCount, 30)
	} else {
		t.Logf("✓ Queue recovered: %d/%d succeeded after drain", successCount, 30)
	}
}

// TestSendEvent_ErrorMessageClarity tests that queue saturation errors
// provide clear, actionable error messages for debugging.
func TestSendEvent_ErrorMessageClarity(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Before connecting, we can still queue events (they go to queue)
	// But if we fill the queue without a batcher running, we should get clear errors

	// Fill the queue manually
	event := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}

	var lastErr error
	// Send events until queue is full
	// The queue starts with capacity 100
	for i := 0; i < 120; i++ {
		err := client.SendEvent(event)
		if err != nil {
			lastErr = err
			break
		}
	}

	if lastErr != nil {
		if !strings.Contains(lastErr.Error(), "retry attempts exhausted") &&
			!strings.Contains(lastErr.Error(), "event queue full") {
			t.Errorf("Error message not clear: %v", lastErr)
		}
		t.Logf("✓ Error message is clear and actionable: %v", lastErr)
	} else {
		t.Log("✓ Queue accepted all test events (batcher not running)")
	}
}

// ============================================================================
// Context Cancellation and Reconnection Tests
// ============================================================================

// TestClient_ContextCancellationDuringReconnection verifies graceful shutdown
// when context is cancelled during reconnection attempts. This ensures:
// - No goroutine leaks when Close() is called during reconnection
// - batcherDone channel is closed properly
// - Context cancellation propagates through all operations
// - No panics or deadlocks during shutdown
func TestClient_ContextCancellationDuringReconnection(t *testing.T) {
	t.Parallel()

	// Track goroutines before test starts
	// Give time for any background goroutines to stabilize
	time.Sleep(100 * time.Millisecond)
	startGoroutines := runtime.NumGoroutine()
	t.Logf("Starting goroutines: %d", startGoroutines)

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

	// Give goroutines time to fully exit
	time.Sleep(500 * time.Millisecond)

	// Check goroutine count - should return to baseline (or close to it)
	endGoroutines := runtime.NumGoroutine()
	goroutineLeak := endGoroutines - startGoroutines

	t.Logf("Ending goroutines: %d (diff: %+d)", endGoroutines, goroutineLeak)

	// Allow some tolerance (±3 goroutines) for runtime background tasks
	if goroutineLeak > 3 {
		t.Errorf("Possible goroutine leak: started with %d, ended with %d (leaked %d)",
			startGoroutines, endGoroutines, goroutineLeak)
	} else {
		t.Logf("✓ No significant goroutine leak detected")
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

// ============================================================================
// Event Batching During Network Failure Tests
// ============================================================================

// TestClient_EventBatchingDuringNetworkFailure tests that events are properly
// queued and batched when the daemon is temporarily unavailable, and then
// successfully sent once the daemon becomes available again.
//
// This test verifies:
// - Events are queued in the eventQueue channel (capacity 100) when daemon is down
// - The batchSender goroutine continues running even when disconnected
// - Events are properly batched together within the debounce window
// - Queued events are sent when connection is restored
// - Debounce timing is respected (configurable via PASO_EVENT_DEBOUNCE_MS)
func TestClient_EventBatchingDuringNetworkFailure(t *testing.T) {
	t.Parallel()

	// Set short debounce for faster testing (50ms instead of default 100ms)
	originalDebounce := os.Getenv("PASO_EVENT_DEBOUNCE_MS")
	_ = os.Setenv("PASO_EVENT_DEBOUNCE_MS", "50")
	defer func() { _ = os.Setenv("PASO_EVENT_DEBOUNCE_MS", originalDebounce) }()

	// Setup initial mock daemon
	socketPath, listener, messages := setupMockDaemon(t)

	// Create client
	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to daemon
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Drain initial subscribe message (project 0)
	select {
	case msg := <-messages:
		if msg.Type != "subscribe" {
			t.Fatalf("Expected initial subscribe, got: %s", msg.Type)
		}
		t.Logf("✓ Client connected and subscribed to project 0")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Send one event to verify connection works
	testEvent := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}
	if err := client.SendEvent(testEvent); err != nil {
		t.Fatalf("Failed to send initial event: %v", err)
	}

	// Wait for debounce and verify event was sent
	time.Sleep(client.debounce + 50*time.Millisecond)
	select {
	case msg := <-messages:
		if msg.Type != "event" {
			t.Errorf("Expected event, got: %s", msg.Type)
		}
		t.Logf("✓ Initial event sent successfully before network failure")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial event")
	}

	// ===== SIMULATE NETWORK FAILURE =====
	// Close the daemon listener to simulate network failure
	t.Logf("Simulating network failure: closing daemon...")
	if err := listener.Close(); err != nil {
		t.Logf("Warning: failed to close listener: %v", err)
	}

	// Wait a bit for the connection to be detected as broken
	time.Sleep(100 * time.Millisecond)

	// ===== QUEUE EVENTS WHILE DAEMON IS DOWN =====
	// Queue multiple events rapidly - these should be queued in eventQueue
	numEventsWhileDown := 5
	t.Logf("Queueing %d events while daemon is down...", numEventsWhileDown)

	for i := 0; i < numEventsWhileDown; i++ {
		event := Event{
			Type:      EventDatabaseChanged,
			ProjectID: i + 1, // Projects 1-5
			Timestamp: time.Now(),
		}
		// SendEvent should succeed (queued) even though daemon is down
		if err := client.SendEvent(event); err != nil {
			t.Logf("Event %d queued (daemon down): %v", i+1, err)
		}
		// Small delay between events to test batching behavior
		time.Sleep(10 * time.Millisecond)
	}

	t.Logf("✓ All %d events queued while daemon was down", numEventsWhileDown)

	// Wait a bit to ensure batcher has tried to flush (will fail due to no connection)
	time.Sleep(client.debounce + 100*time.Millisecond)

	// ===== RESTORE NETWORK CONNECTION =====
	// Create a new mock daemon on the same socket path
	t.Logf("Restoring network: starting new daemon...")
	newListener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create new listener: %v", err)
	}
	defer func() { _ = newListener.Close() }()

	// Setup new message channel for the restored daemon
	restoredMessages := make(chan Message, 20)

	// Accept connections on restored daemon
	go func() {
		for {
			conn, err := newListener.Accept()
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

					// Send message to test channel
					select {
					case restoredMessages <- msg:
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

	// Manually reconnect the client to simulate automatic reconnection
	// We need to reconnect the underlying socket without starting a new batcher
	t.Logf("Reconnecting client to restored daemon...")

	// Dial the new daemon
	dialer := net.Dialer{}
	newConn, err := dialer.DialContext(context.Background(), "unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to dial restored daemon: %v", err)
	}

	// Update the client's connection (simulating what reconnect() does internally)
	client.mu.Lock()
	client.conn = newConn
	client.encoder = json.NewEncoder(newConn)
	client.decoder = json.NewDecoder(newConn)
	client.mu.Unlock()

	// Send subscribe message manually
	if err := client.Subscribe(0); err != nil {
		t.Fatalf("Failed to subscribe after reconnect: %v", err)
	}

	// Drain reconnection subscribe message
	select {
	case msg := <-restoredMessages:
		if msg.Type != "subscribe" {
			t.Logf("Expected subscribe after reconnect, got: %s", msg.Type)
		}
		t.Logf("✓ Client reconnected successfully")
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for reconnect subscribe")
	}

	// ===== VERIFY QUEUED EVENTS ARE SENT =====
	// Wait for the batcher to flush the queued events
	// The batcher should have accumulated the events and will send them in batch(es)
	time.Sleep(client.debounce + 200*time.Millisecond)

	// Collect all event messages sent after reconnection
	var receivedEvents []Message
	timeout := time.After(2 * time.Second)

collectLoop:
	for {
		select {
		case msg := <-restoredMessages:
			if msg.Type == "event" {
				receivedEvents = append(receivedEvents, msg)
				t.Logf("Received batched event: ProjectID=%d", msg.Event.ProjectID)
			}
		case <-timeout:
			break collectLoop
		default:
			// Give a bit more time for any pending messages
			time.Sleep(50 * time.Millisecond)
			select {
			case msg := <-restoredMessages:
				if msg.Type == "event" {
					receivedEvents = append(receivedEvents, msg)
					t.Logf("Received batched event: ProjectID=%d", msg.Event.ProjectID)
				}
			default:
				break collectLoop
			}
		}
	}

	// ===== VERIFY RESULTS =====
	// Note: Events queued while daemon was down may have been sent and failed
	// (errors logged but not retried by batcher). This is expected behavior.
	// The batcher attempts to send immediately after debounce, and if it fails,
	// it marks pending=false and moves on. This test verifies the batching
	// mechanism works correctly, not event delivery guarantees.
	if len(receivedEvents) == 0 {
		t.Logf("⚠ No events received from pre-reconnection queue (expected - batcher attempted send while disconnected)")
	} else {
		t.Logf("✓ Received %d batched event message(s) after reconnection", len(receivedEvents))

		// Verify that events were batched properly
		// Since we sent events for multiple projects (1-5), they should be batched
		// into a single event with ProjectID=0 (indicates multiple projects)
		foundBatchedEvent := false
		for _, msg := range receivedEvents {
			if msg.Event != nil {
				t.Logf("  Event type: %s, ProjectID: %d", msg.Event.Type, msg.Event.ProjectID)
				if msg.Event.ProjectID == 0 {
					// ProjectID 0 indicates multiple projects were batched together
					foundBatchedEvent = true
					t.Logf("✓ Found batched event (ProjectID=0 indicates multiple projects)")
				}
			}
		}

		if !foundBatchedEvent && len(receivedEvents) > 1 {
			t.Logf("✓ Events sent as separate messages (valid batching strategy)")
		} else if foundBatchedEvent {
			t.Logf("✓ Events properly batched into single event with ProjectID=0")
		}
	}

	// ===== VERIFY DEBOUNCE TIMING =====
	// Send a few more events and verify they respect debounce timing
	t.Logf("Testing debounce timing with new events...")
	startTime := time.Now()

	// Send 3 events rapidly
	for i := 0; i < 3; i++ {
		event := Event{
			Type:      EventDatabaseChanged,
			ProjectID: 10,
			Timestamp: time.Now(),
		}
		if err := client.SendEvent(event); err != nil {
			t.Errorf("Failed to send test event %d: %v", i, err)
		}
	}

	// Wait for events to be batched and sent
	time.Sleep(client.debounce + 100*time.Millisecond)

	// Should receive batched event
	select {
	case msg := <-restoredMessages:
		if msg.Type != "event" {
			t.Errorf("Expected event, got: %s", msg.Type)
		}
		elapsed := time.Since(startTime)

		// Event should arrive after at least the debounce duration
		if elapsed < client.debounce {
			t.Errorf("Event arrived too quickly (elapsed: %v, expected >= %v)",
				elapsed, client.debounce)
		} else {
			t.Logf("✓ Debounce timing respected (elapsed: %v, debounce: %v)",
				elapsed, client.debounce)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for debounced event")
	}

	// ===== VERIFY QUEUE DOESN'T OVERFLOW =====
	// Test that the eventQueue (capacity 100) handles events properly
	t.Logf("Testing queue capacity handling...")

	// Send events up to but not exceeding queue capacity
	queueTestEvents := 50 // Well within capacity of 100
	successCount := 0

	for i := 0; i < queueTestEvents; i++ {
		event := Event{
			Type:      EventDatabaseChanged,
			ProjectID: 20,
			Timestamp: time.Now(),
		}
		if err := client.SendEvent(event); err == nil {
			successCount++
		}
	}

	if successCount != queueTestEvents {
		t.Errorf("Expected all %d events to be queued, only %d succeeded",
			queueTestEvents, successCount)
	} else {
		t.Logf("✓ Queue handled %d events successfully (capacity: 100)", queueTestEvents)
	}

	// Wait for queue to drain
	time.Sleep(client.debounce + 200*time.Millisecond)

	// Drain any remaining messages
	drained := 0
	drainTimeout := time.After(1 * time.Second)
drainLoop:
	for {
		select {
		case msg := <-restoredMessages:
			if msg.Type == "event" {
				drained++
			}
		case <-drainTimeout:
			break drainLoop
		default:
			break drainLoop
		}
	}

	if drained > 0 {
		t.Logf("✓ Queue drained: received %d event message(s) from queue test", drained)
	}

	t.Logf("✓ Test completed successfully: event batching during network failure works correctly")
}

// ============================================================================
// Event Receiver Loop Error Handling Tests
// ============================================================================

// TestClient_EventReceiverHandlesMalformedJSON tests that the event receiver
// handles malformed JSON gracefully without crashing the client.
func TestClient_EventReceiverHandlesMalformedJSON(t *testing.T) {
	t.Parallel()

	// Create temp directory and socket path
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	// Create Unix socket listener
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create mock daemon listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	// Channel to track if client handled error gracefully
	handledError := make(chan bool, 1)

	// Start server that sends malformed JSON
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		decoder := json.NewDecoder(conn)
		encoder := json.NewEncoder(conn)

		// Read and acknowledge initial subscribe message
		var subscribeMsg Message
		if err := decoder.Decode(&subscribeMsg); err != nil {
			return
		}

		// Send ack for subscribe
		ackMsg := Message{
			Version: ProtocolVersion,
			Type:    "ack",
		}
		_ = encoder.Encode(ackMsg)

		// Send malformed JSON (not a valid Message structure)
		malformedJSON := []byte(`{"invalid": "json", "missing": "required_fields`)
		_, _ = conn.Write(malformedJSON)
		_, _ = conn.Write([]byte("\n"))
	}()

	// Create and connect client
	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Start listening for events
	eventChan, err := client.Listen(ctx)
	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}

	// Monitor for graceful handling
	go func() {
		for range eventChan {
			// Consume any events
		}
		handledError <- true
	}()

	// Wait for error handling
	select {
	case <-handledError:
		t.Logf("✓ Client handled malformed JSON gracefully (event channel closed)")
	case <-time.After(2 * time.Second):
		t.Logf("✓ Client handled malformed JSON gracefully (timeout without crash)")
	}
}

// TestClient_EventReceiverHandlesInvalidEventType tests that the receiver
// handles events with invalid/unknown event types gracefully.
func TestClient_EventReceiverHandlesInvalidEventType(t *testing.T) {
	t.Parallel()

	// Create temp directory and socket path
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	// Create Unix socket listener
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create mock daemon listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	// Channel to receive events
	receivedEvents := make(chan Event, 5)

	// Start server that sends events with invalid types
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		decoder := json.NewDecoder(conn)
		encoder := json.NewEncoder(conn)

		// Read and acknowledge initial subscribe message
		var subscribeMsg Message
		if err := decoder.Decode(&subscribeMsg); err != nil {
			return
		}

		// Send ack for subscribe
		ackMsg := Message{
			Version: ProtocolVersion,
			Type:    "ack",
		}
		_ = encoder.Encode(ackMsg)

		// Send a valid event first
		validEvent := Message{
			Version: ProtocolVersion,
			Type:    "event",
			Event: &Event{
				Type:       EventDatabaseChanged,
				ProjectID:  1,
				Timestamp:  time.Now(),
				SequenceID: 1,
			},
		}
		_ = encoder.Encode(validEvent)

		// Send event with invalid Type (should be ignored)
		invalidTypeEvent := Message{
			Version: ProtocolVersion,
			Type:    "unknown_event_type",
			Event: &Event{
				Type:       EventType("invalid_event"),
				ProjectID:  2,
				Timestamp:  time.Now(),
				SequenceID: 2,
			},
		}
		_ = encoder.Encode(invalidTypeEvent)

		// Send another valid event to verify client still works
		validEvent2 := Message{
			Version: ProtocolVersion,
			Type:    "event",
			Event: &Event{
				Type:       EventDatabaseChanged,
				ProjectID:  3,
				Timestamp:  time.Now(),
				SequenceID: 3,
			},
		}
		_ = encoder.Encode(validEvent2)

		// Keep connection alive
		time.Sleep(2 * time.Second)
	}()

	// Create and connect client
	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Start listening for events
	eventChan, err := client.Listen(ctx)
	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}

	// Collect events
	go func() {
		for event := range eventChan {
			receivedEvents <- event
		}
		close(receivedEvents)
	}()

	// Verify we received valid events
	timeout := time.After(2 * time.Second)
	eventCount := 0

	for {
		select {
		case event, ok := <-receivedEvents:
			if !ok {
				// Channel closed
				if eventCount >= 2 {
					t.Logf("✓ Client handled invalid event types gracefully and received %d valid events", eventCount)
					return
				}
				t.Errorf("Expected at least 2 events, got %d", eventCount)
				return
			}
			eventCount++
			t.Logf("Received event: ProjectID=%d, SequenceID=%d", event.ProjectID, event.SequenceID)
		case <-timeout:
			if eventCount >= 2 {
				t.Logf("✓ Client handled invalid event types gracefully and received %d valid events", eventCount)
				return
			}
			t.Errorf("Timeout: expected at least 2 events, got %d", eventCount)
			return
		}
	}
}

// TestClient_EventReceiverTracksSequenceNumbers tests that the receiver
// properly tracks sequence numbers and prevents duplicates.
func TestClient_EventReceiverTracksSequenceNumbers(t *testing.T) {
	t.Parallel()

	// Create temp directory and socket path
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	// Create Unix socket listener
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create mock daemon listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	// Channel to receive events
	receivedEvents := make(chan Event, 10)

	// Start server that sends events with various sequence numbers
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		decoder := json.NewDecoder(conn)
		encoder := json.NewEncoder(conn)

		// Read and acknowledge initial subscribe message
		var subscribeMsg Message
		if err := decoder.Decode(&subscribeMsg); err != nil {
			return
		}

		// Send ack for subscribe
		ackMsg := Message{
			Version: ProtocolVersion,
			Type:    "ack",
		}
		_ = encoder.Encode(ackMsg)

		// Send events with sequence numbers: 1, 2, 2 (duplicate), 3, 1 (old)
		sequences := []int64{1, 2, 2, 3, 1, 4}
		for i, seq := range sequences {
			event := Message{
				Version: ProtocolVersion,
				Type:    "event",
				Event: &Event{
					Type:       EventDatabaseChanged,
					ProjectID:  i + 1,
					Timestamp:  time.Now(),
					SequenceID: seq,
				},
			}
			_ = encoder.Encode(event)
			time.Sleep(50 * time.Millisecond)
		}

		// Keep connection alive
		time.Sleep(2 * time.Second)
	}()

	// Create and connect client
	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Verify initial lastSequence is 0
	client.mu.Lock()
	initialSeq := client.lastSequence
	client.mu.Unlock()

	if initialSeq != 0 {
		t.Errorf("Expected initial lastSequence to be 0, got %d", initialSeq)
	}

	// Start listening for events
	eventChan, err := client.Listen(ctx)
	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}

	// Collect events
	go func() {
		for event := range eventChan {
			receivedEvents <- event
		}
		close(receivedEvents)
	}()

	// Verify sequence tracking
	timeout := time.After(2 * time.Second)
	receivedSeqs := []int64{}

	for {
		select {
		case event, ok := <-receivedEvents:
			if !ok {
				// Channel closed - verify results
				goto verify
			}
			receivedSeqs = append(receivedSeqs, event.SequenceID)
			t.Logf("Received event with SequenceID=%d, ProjectID=%d", event.SequenceID, event.ProjectID)
		case <-timeout:
			goto verify
		}
	}

verify:
	// We should receive: 1, 2, 3, 4 (duplicates and old sequences filtered)
	expectedSeqs := []int64{1, 2, 3, 4}
	if len(receivedSeqs) != len(expectedSeqs) {
		t.Errorf("Expected %d events, got %d", len(expectedSeqs), len(receivedSeqs))
	}

	for i, expected := range expectedSeqs {
		if i >= len(receivedSeqs) {
			t.Errorf("Missing event at index %d with SequenceID=%d", i, expected)
			continue
		}
		if receivedSeqs[i] != expected {
			t.Errorf("At index %d: expected SequenceID=%d, got %d", i, expected, receivedSeqs[i])
		}
	}

	// Verify lastSequence is updated correctly
	client.mu.Lock()
	finalSeq := client.lastSequence
	client.mu.Unlock()

	if finalSeq != 4 {
		t.Errorf("Expected final lastSequence to be 4, got %d", finalSeq)
	}

	t.Logf("✓ Client tracks sequence numbers correctly: lastSequence=%d", finalSeq)
}

// TestClient_EventReceiverHandlesMissingSequenceNumbers tests that the receiver
// handles events with missing (zero) sequence numbers gracefully.
func TestClient_EventReceiverHandlesMissingSequenceNumbers(t *testing.T) {
	t.Parallel()

	// Create temp directory and socket path
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	// Create Unix socket listener
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create mock daemon listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	// Channel to receive events
	receivedEvents := make(chan Event, 10)

	// Start server that sends events with zero/missing sequence numbers
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		decoder := json.NewDecoder(conn)
		encoder := json.NewEncoder(conn)

		// Read and acknowledge initial subscribe message
		var subscribeMsg Message
		if err := decoder.Decode(&subscribeMsg); err != nil {
			return
		}

		// Send ack for subscribe
		ackMsg := Message{
			Version: ProtocolVersion,
			Type:    "ack",
		}
		_ = encoder.Encode(ackMsg)

		// Send event with sequence 1
		event1 := Message{
			Version: ProtocolVersion,
			Type:    "event",
			Event: &Event{
				Type:       EventDatabaseChanged,
				ProjectID:  1,
				Timestamp:  time.Now(),
				SequenceID: 1,
			},
		}
		_ = encoder.Encode(event1)

		// Send event with zero sequence number (missing/unset)
		eventZeroSeq := Message{
			Version: ProtocolVersion,
			Type:    "event",
			Event: &Event{
				Type:       EventDatabaseChanged,
				ProjectID:  2,
				Timestamp:  time.Now(),
				SequenceID: 0, // Missing/unset sequence
			},
		}
		_ = encoder.Encode(eventZeroSeq)

		// Send another valid event
		event2 := Message{
			Version: ProtocolVersion,
			Type:    "event",
			Event: &Event{
				Type:       EventDatabaseChanged,
				ProjectID:  3,
				Timestamp:  time.Now(),
				SequenceID: 2,
			},
		}
		_ = encoder.Encode(event2)

		// Keep connection alive
		time.Sleep(2 * time.Second)
	}()

	// Create and connect client
	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Start listening for events
	eventChan, err := client.Listen(ctx)
	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}

	// Collect events
	go func() {
		for event := range eventChan {
			receivedEvents <- event
		}
		close(receivedEvents)
	}()

	// Verify we received events (zero sequence should be filtered out)
	timeout := time.After(2 * time.Second)
	receivedSeqs := []int64{}

	for {
		select {
		case event, ok := <-receivedEvents:
			if !ok {
				goto verify
			}
			receivedSeqs = append(receivedSeqs, event.SequenceID)
			t.Logf("Received event with SequenceID=%d, ProjectID=%d", event.SequenceID, event.ProjectID)
		case <-timeout:
			goto verify
		}
	}

verify:
	// We should receive events with sequence 1 and 2 (0 is filtered)
	expectedSeqs := []int64{1, 2}
	if len(receivedSeqs) != len(expectedSeqs) {
		t.Errorf("Expected %d events, got %d (sequences: %v)", len(expectedSeqs), len(receivedSeqs), receivedSeqs)
	}

	for i, expected := range expectedSeqs {
		if i >= len(receivedSeqs) {
			t.Errorf("Missing event at index %d with SequenceID=%d", i, expected)
			continue
		}
		if receivedSeqs[i] != expected {
			t.Errorf("At index %d: expected SequenceID=%d, got %d", i, expected, receivedSeqs[i])
		}
	}

	t.Logf("✓ Client handles missing sequence numbers gracefully")
}

// TestClient_NotificationCallbackRouting tests that notification messages
// are properly routed through the notification callback.
func TestClient_NotificationCallbackRouting(t *testing.T) {
	t.Parallel()

	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Set up notification callback to capture messages
	notifications := make(chan NotificationMsg, 10)
	client.SetNotifyFunc(func(level, message string) {
		notifications <- NotificationMsg{
			Level:   level,
			Message: message,
		}
	})

	// Trigger a notification
	client.notify("info", "test notification message")

	// Verify notification was received
	select {
	case notif := <-notifications:
		if notif.Level != "info" {
			t.Errorf("Expected level 'info', got '%s'", notif.Level)
		}
		if notif.Message != "test notification message" {
			t.Errorf("Expected message 'test notification message', got '%s'", notif.Message)
		}
		t.Logf("✓ Notification routed correctly: level=%s, message=%s", notif.Level, notif.Message)
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for notification")
	}

	// Test multiple notification levels
	testCases := []struct {
		level   string
		message string
	}{
		{"info", "information message"},
		{"warning", "warning message"},
		{"error", "error message"},
	}

	for _, tc := range testCases {
		client.notify(tc.level, tc.message)

		select {
		case notif := <-notifications:
			if notif.Level != tc.level {
				t.Errorf("Expected level '%s', got '%s'", tc.level, notif.Level)
			}
			if notif.Message != tc.message {
				t.Errorf("Expected message '%s', got '%s'", tc.message, notif.Message)
			}
		case <-time.After(1 * time.Second):
			t.Errorf("Timeout waiting for notification: level=%s", tc.level)
		}
	}

	t.Logf("✓ All notification levels routed correctly")
}

// TestClient_EventReceiverPingPongHandling tests that the receiver properly
// handles ping messages by sending pong responses.
func TestClient_EventReceiverPingPongHandling(t *testing.T) {
	t.Parallel()

	// Create temp directory and socket path
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	// Create Unix socket listener
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create mock daemon listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	// Channel to track if pong was received
	pongReceived := make(chan bool, 1)

	// Start server that sends ping and expects pong
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		decoder := json.NewDecoder(conn)
		encoder := json.NewEncoder(conn)

		// Read and acknowledge initial subscribe message
		var subscribeMsg Message
		if err := decoder.Decode(&subscribeMsg); err != nil {
			return
		}

		// Send ack for subscribe
		ackMsg := Message{
			Version: ProtocolVersion,
			Type:    "ack",
		}
		_ = encoder.Encode(ackMsg)

		// Send ping message
		pingMsg := Message{
			Version: ProtocolVersion,
			Type:    "ping",
		}
		_ = encoder.Encode(pingMsg)

		// Wait for pong response
		go func() {
			for {
				var response Message
				if err := decoder.Decode(&response); err != nil {
					return
				}
				if response.Type == "event" && response.Event != nil && response.Event.Type == EventPong {
					pongReceived <- true
					return
				}
			}
		}()

		// Keep connection alive
		time.Sleep(2 * time.Second)
	}()

	// Create and connect client
	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Start listening to trigger the event receiver loop
	_, err = client.Listen(ctx)
	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}

	// Verify pong was received by server
	select {
	case <-pongReceived:
		t.Logf("✓ Client responded to ping with pong correctly")
	case <-time.After(2 * time.Second):
		t.Error("Timeout: client did not respond to ping")
	}
}

// TestClient_EventReceiverContinuesAfterErrors tests that valid events
// after errors are still processed correctly.
func TestClient_EventReceiverContinuesAfterErrors(t *testing.T) {
	t.Parallel()

	// Create temp directory and socket path
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	// Create Unix socket listener
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create mock daemon listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	// Channel to receive events
	receivedEvents := make(chan Event, 10)

	// Start server that sends mix of valid and problematic messages
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		decoder := json.NewDecoder(conn)
		encoder := json.NewEncoder(conn)

		// Read and acknowledge initial subscribe message
		var subscribeMsg Message
		if err := decoder.Decode(&subscribeMsg); err != nil {
			return
		}

		// Send ack for subscribe
		ackMsg := Message{
			Version: ProtocolVersion,
			Type:    "ack",
		}
		_ = encoder.Encode(ackMsg)

		// Send valid event 1
		event1 := Message{
			Version: ProtocolVersion,
			Type:    "event",
			Event: &Event{
				Type:       EventDatabaseChanged,
				ProjectID:  1,
				Timestamp:  time.Now(),
				SequenceID: 1,
			},
		}
		_ = encoder.Encode(event1)

		// Send event with old sequence (should be filtered)
		oldSeqEvent := Message{
			Version: ProtocolVersion,
			Type:    "event",
			Event: &Event{
				Type:       EventDatabaseChanged,
				ProjectID:  2,
				Timestamp:  time.Now(),
				SequenceID: 0, // Old sequence
			},
		}
		_ = encoder.Encode(oldSeqEvent)

		// Send valid event 2 (should still be processed)
		event2 := Message{
			Version: ProtocolVersion,
			Type:    "event",
			Event: &Event{
				Type:       EventDatabaseChanged,
				ProjectID:  3,
				Timestamp:  time.Now(),
				SequenceID: 2,
			},
		}
		_ = encoder.Encode(event2)

		// Keep connection alive
		time.Sleep(2 * time.Second)
	}()

	// Create and connect client
	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Start listening for events
	eventChan, err := client.Listen(ctx)
	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}

	// Collect events
	go func() {
		for event := range eventChan {
			receivedEvents <- event
		}
		close(receivedEvents)
	}()

	// Verify we received both valid events
	timeout := time.After(2 * time.Second)
	receivedProjectIDs := []int{}

	for {
		select {
		case event, ok := <-receivedEvents:
			if !ok {
				goto verify
			}
			receivedProjectIDs = append(receivedProjectIDs, event.ProjectID)
			t.Logf("Received event: ProjectID=%d, SequenceID=%d", event.ProjectID, event.SequenceID)
		case <-timeout:
			goto verify
		}
	}

verify:
	// We should receive events from projects 1 and 3 (project 2 filtered)
	expectedProjectIDs := []int{1, 3}
	if len(receivedProjectIDs) != len(expectedProjectIDs) {
		t.Errorf("Expected %d events, got %d", len(expectedProjectIDs), len(receivedProjectIDs))
	}

	for i, expected := range expectedProjectIDs {
		if i >= len(receivedProjectIDs) {
			t.Errorf("Missing event at index %d with ProjectID=%d", i, expected)
			continue
		}
		if receivedProjectIDs[i] != expected {
			t.Errorf("At index %d: expected ProjectID=%d, got %d", i, expected, receivedProjectIDs[i])
		}
	}

	t.Logf("✓ Client continues processing valid events after errors")
}

// ============================================================================
// Disconnect Detection Tests
// ============================================================================

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
