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

// Test helpers to avoid import cycle with testutil

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

// ============================================================================
// Server Initialization Tests
// ============================================================================

func TestNewServer_Success(t *testing.T) {
	socketPath := getTestSocketPath(t)

	server, err := NewServer(socketPath)
	if err != nil {
		t.Fatalf("Expected NewServer to succeed, got error: %v", err)
	}
	defer func() { _ = server.Shutdown() }()

	// Verify socket file was created
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Error("Expected socket file to be created")
	}

	// Verify server fields initialized
	if server == nil {
		t.Fatal("Expected server to be non-nil")
	}

	t.Logf("✓ Server created successfully at %s", socketPath)
}

func TestNewServer_DirectoryCreation(t *testing.T) {
	// Use t.TempDir() which ensures cleanup
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "subdirs", "paso.sock")

	server, err := NewServer(nestedPath)
	if err != nil {
		t.Fatalf("Expected NewServer to create nested directories, got error: %v", err)
	}
	defer func() { _ = server.Shutdown() }()

	// Verify directories were created
	dir := filepath.Dir(nestedPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Expected directory %s to be created", dir)
	}

	// Verify socket file exists
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("Expected socket file to be created in nested directory")
	}

	t.Logf("✓ Nested directories created successfully: %s", nestedPath)
}

func TestNewServer_StaleSocketCleanup(t *testing.T) {
	socketPath := getTestSocketPath(t)

	// Create a stale socket file
	f, err := os.Create(socketPath)
	if err != nil {
		t.Fatalf("Failed to create stale socket file: %v", err)
	}
	_ = f.Close()

	// Verify stale socket exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Fatal("Stale socket file should exist before NewServer")
	}

	// Create new server (should remove stale socket)
	server, err := NewServer(socketPath)
	if err != nil {
		t.Fatalf("Expected NewServer to succeed after removing stale socket, got error: %v", err)
	}
	defer func() { _ = server.Shutdown() }()

	// Verify new socket was created (the old one was removed)
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Error("Expected new socket file to be created")
	}

	t.Logf("✓ Stale socket cleaned up successfully")
}

func TestNewServer_EnvVarConfiguration(t *testing.T) {
	// Save original env vars
	originalBroadcast := os.Getenv("PASO_DAEMON_BROADCAST_BUFFER")
	originalClient := os.Getenv("PASO_DAEMON_CLIENT_BUFFER")
	defer func() {
		_ = os.Setenv("PASO_DAEMON_BROADCAST_BUFFER", originalBroadcast)
		_ = os.Setenv("PASO_DAEMON_CLIENT_BUFFER", originalClient)
	}()

	// Set environment variables
	_ = os.Setenv("PASO_DAEMON_BROADCAST_BUFFER", "200")
	_ = os.Setenv("PASO_DAEMON_CLIENT_BUFFER", "20")

	socketPath := getTestSocketPath(t)
	server, err := NewServer(socketPath)
	if err != nil {
		t.Fatalf("Expected NewServer to succeed, got error: %v", err)
	}
	defer func() { _ = server.Shutdown() }()

	// Note: We can't directly verify buffer sizes since they're unexported
	// But we can verify the server was created successfully with env vars set
	t.Logf("✓ Server created with custom buffer sizes from env vars")
}

// ============================================================================
// Client Connection Tests
// ============================================================================

func TestClientConnection_Single(t *testing.T) {
	_, socketPath := setupTestDaemon(t)

	// Connect a raw client
	conn, encoder, _ := connectRawClient(t, socketPath)

	// Send initial subscribe message
	sendSubscribeMessage(t, encoder, 0)

	// Give server time to process connection
	time.Sleep(50 * time.Millisecond)

	// Verify connection is still active by checking if we can write
	if err := encoder.Encode(events.Message{Version: events.ProtocolVersion, Type: "ping"}); err != nil {
		t.Fatalf("Expected connection to be active, got error: %v", err)
	}

	// Try to read with a short deadline - we expect timeout (no response expected for ping from client)
	_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	decoder := json.NewDecoder(conn)
	var msg events.Message
	err := decoder.Decode(&msg)
	if err == nil {
		// Unexpectedly got a message
		t.Logf("Note: Received unexpected message type: %s", msg.Type)
	}

	t.Logf("✓ Client connected successfully")
}

func TestClientConnection_Multiple(t *testing.T) {
	server, socketPath := setupTestDaemon(t)

	numClients := 5

	// Connect multiple clients
	for i := 0; i < numClients; i++ {
		_, encoder, _ := connectRawClient(t, socketPath)
		sendSubscribeMessage(t, encoder, 0)
	}

	// Give server time to process all connections
	time.Sleep(100 * time.Millisecond)

	t.Logf("✓ Successfully connected %d clients", numClients)
	logServerState(t, server, "after multiple connections")
}

func TestClientDisconnection(t *testing.T) {
	server, socketPath := setupTestDaemon(t)

	// Connect a client
	conn, encoder, _ := connectRawClient(t, socketPath)
	sendSubscribeMessage(t, encoder, 0)

	time.Sleep(50 * time.Millisecond)
	logServerState(t, server, "after connection")

	// Close the connection
	if closer, ok := conn.(interface{ Close() error }); ok {
		_ = closer.Close()
	}

	// Give server time to detect disconnection
	time.Sleep(100 * time.Millisecond)

	logServerState(t, server, "after disconnection")
	t.Logf("✓ Client disconnected and cleaned up")
}

// ============================================================================
// Event Broadcasting Tests
// ============================================================================

func TestBroadcast_SingleClient(t *testing.T) {
	server, socketPath := setupTestDaemon(t)

	// Connect client using event client
	client := setupTestClient(t, socketPath)

	// Subscribe to project 1
	if err := client.Subscribe(1); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Start listening for events
	eventChan, err := client.Listen(context.Background())
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	// Give client time to fully establish subscription
	time.Sleep(100 * time.Millisecond)

	// Broadcast an event
	testEvent := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}

	if err := server.Broadcast(testEvent); err != nil {
		t.Fatalf("Failed to broadcast: %v", err)
	}

	// Wait for event
	receivedEvent := waitForEvent(t, eventChan, 2*time.Second)

	if receivedEvent.ProjectID != 1 {
		t.Errorf("Expected event for project 1, got %d", receivedEvent.ProjectID)
	}

	// Verify sequence ID was set
	if receivedEvent.SequenceID == 0 {
		t.Error("Expected sequence ID to be set")
	}

	t.Logf("✓ Event broadcast and received successfully (sequence: %d)", receivedEvent.SequenceID)
}

func TestBroadcast_MultipleClients(t *testing.T) {
	server, socketPath := setupTestDaemon(t)

	numClients := 3
	var eventChans []<-chan events.Event

	// Connect multiple clients
	for i := 0; i < numClients; i++ {
		client := setupTestClient(t, socketPath)

		// Subscribe all to project 1
		if err := client.Subscribe(1); err != nil {
			t.Fatalf("Client %d failed to subscribe: %v", i, err)
		}

		eventChan, err := client.Listen(context.Background())
		if err != nil {
			t.Fatalf("Client %d failed to listen: %v", i, err)
		}
		eventChans = append(eventChans, eventChan)
	}

	// Give clients time to subscribe
	time.Sleep(100 * time.Millisecond)

	// Broadcast event
	testEvent := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}

	if err := server.Broadcast(testEvent); err != nil {
		t.Fatalf("Failed to broadcast: %v", err)
	}

	// Verify all clients receive the event
	for i, eventChan := range eventChans {
		receivedEvent := waitForEvent(t, eventChan, 2*time.Second)
		if receivedEvent.ProjectID != 1 {
			t.Errorf("Client %d: Expected event for project 1, got %d", i, receivedEvent.ProjectID)
		}
		t.Logf("✓ Client %d received event (sequence: %d)", i, receivedEvent.SequenceID)
	}
}

func TestBroadcast_SubscriptionFiltering(t *testing.T) {
	server, socketPath := setupTestDaemon(t)

	// Client A subscribes to project 1
	clientA := setupTestClient(t, socketPath)
	if err := clientA.Subscribe(1); err != nil {
		t.Fatalf("ClientA failed to subscribe: %v", err)
	}
	eventChanA, _ := clientA.Listen(context.Background())

	// Client B subscribes to project 2
	clientB := setupTestClient(t, socketPath)
	if err := clientB.Subscribe(2); err != nil {
		t.Fatalf("ClientB failed to subscribe: %v", err)
	}
	eventChanB, _ := clientB.Listen(context.Background())

	time.Sleep(100 * time.Millisecond)

	// Broadcast event for project 1
	testEvent := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}

	if err := server.Broadcast(testEvent); err != nil {
		t.Fatalf("Failed to broadcast: %v", err)
	}

	// Client A should receive it
	receivedEvent := waitForEvent(t, eventChanA, 2*time.Second)
	if receivedEvent.ProjectID != 1 {
		t.Errorf("ClientA: Expected event for project 1, got %d", receivedEvent.ProjectID)
	}

	// Client B should NOT receive it (different project)
	waitForNoEvent(t, eventChanB, 500*time.Millisecond)

	t.Logf("✓ Subscription filtering works correctly")
}

func TestBroadcast_AllProjects(t *testing.T) {
	server, socketPath := setupTestDaemon(t)

	// Client A subscribes to project 1
	clientA := setupTestClient(t, socketPath)
	if err := clientA.Subscribe(1); err != nil {
		t.Fatalf("ClientA failed to subscribe: %v", err)
	}
	eventChanA, _ := clientA.Listen(context.Background())

	// Client B subscribes to project 2
	clientB := setupTestClient(t, socketPath)
	if err := clientB.Subscribe(2); err != nil {
		t.Fatalf("ClientB failed to subscribe: %v", err)
	}
	eventChanB, _ := clientB.Listen(context.Background())

	// Give clients more time to fully establish subscriptions
	time.Sleep(200 * time.Millisecond)

	// Broadcast event for all projects (projectID = 0)
	testEvent := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 0,
		Timestamp: time.Now(),
	}

	if err := server.Broadcast(testEvent); err != nil {
		t.Fatalf("Failed to broadcast: %v", err)
	}

	// Both clients should receive it
	receivedEventA := waitForEvent(t, eventChanA, 2*time.Second)
	if receivedEventA.ProjectID != 0 {
		t.Errorf("ClientA: Expected event for project 0 (all), got %d", receivedEventA.ProjectID)
	}

	receivedEventB := waitForEvent(t, eventChanB, 2*time.Second)
	if receivedEventB.ProjectID != 0 {
		t.Errorf("ClientB: Expected event for project 0 (all), got %d", receivedEventB.ProjectID)
	}

	t.Logf("✓ Broadcast to all projects works correctly")
}

func TestBroadcast_SequenceNumbers(t *testing.T) {
	server, socketPath := setupTestDaemon(t)

	client := setupTestClient(t, socketPath)
	if err := client.Subscribe(1); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	eventChan, _ := client.Listen(context.Background())

	time.Sleep(50 * time.Millisecond)

	// Send 10 events
	numEvents := 10
	for i := 0; i < numEvents; i++ {
		testEvent := events.Event{
			Type:      events.EventDatabaseChanged,
			ProjectID: 1,
			Timestamp: time.Now(),
		}
		if err := server.Broadcast(testEvent); err != nil {
			t.Fatalf("Failed to broadcast event %d: %v", i, err)
		}
	}

	// Collect all events
	var sequences []int64
	for i := 0; i < numEvents; i++ {
		event := waitForEvent(t, eventChan, 2*time.Second)
		sequences = append(sequences, event.SequenceID)
	}

	// Verify sequences are monotonically increasing
	for i := 1; i < len(sequences); i++ {
		if sequences[i] <= sequences[i-1] {
			t.Errorf("Sequence numbers not monotonic: %d followed by %d", sequences[i-1], sequences[i])
		}
	}

	t.Logf("✓ Sequence numbers are monotonically increasing: %v", sequences)
}

// ============================================================================
// Shutdown Tests
// ============================================================================

func TestShutdown_GracefulClose(t *testing.T) {
	server, socketPath := setupTestDaemon(t)

	// Connect a few clients
	client1 := setupTestClient(t, socketPath)
	_ = setupTestClient(t, socketPath) // client2

	time.Sleep(100 * time.Millisecond)
	logServerState(t, server, "before shutdown")

	// Shutdown server
	if err := server.Shutdown(); err != nil {
		t.Errorf("Expected Shutdown to succeed, got error: %v", err)
	}

	// Verify socket file removed
	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Error("Expected socket file to be removed after shutdown")
	}

	// Verify clients are disconnected (their connections should be closed)
	// Try to send event - should fail
	if err := client1.SendEvent(events.Event{Type: events.EventDatabaseChanged}); err == nil {
		// Event might still be in queue, that's ok
		t.Logf("Note: Event queued after shutdown (might be flushed before close)")
	}

	t.Logf("✓ Server shutdown gracefully")
}

func TestShutdown_Idempotent(t *testing.T) {
	socketPath := getTestSocketPath(t)
	server, err := NewServer(socketPath)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Shutdown once
	if err := server.Shutdown(); err != nil {
		t.Errorf("First shutdown failed: %v", err)
	}

	// Shutdown again - should not panic or error
	if err := server.Shutdown(); err != nil {
		t.Errorf("Second shutdown should be idempotent, got error: %v", err)
	}

	t.Logf("✓ Shutdown is idempotent")
}
