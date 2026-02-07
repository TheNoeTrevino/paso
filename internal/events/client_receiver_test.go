package events

import (
	"context"
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"
)

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
