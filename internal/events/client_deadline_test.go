package events

import (
	"context"
	"testing"
	"time"
)

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
