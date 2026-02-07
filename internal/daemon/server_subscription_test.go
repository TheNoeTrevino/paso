package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/thenoetrevino/paso/internal/events"
)

// TestSubscription_Switching tests that a client can switch between project
// subscriptions and only receives events for the currently subscribed project.
// This validates the server's subscription routing logic.
//
// Bug context: This test ensures that when switching projects in the TUI,
// the daemon correctly updates which events are routed to the client.
func TestSubscription_Switching(t *testing.T) {
	server, socketPath := setupTestDaemon(t)
	logServerState(t, server, "initial")

	// Create a client and subscribe to project 1
	client := setupTestClient(t, socketPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventChan, err := client.Listen(ctx)
	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}

	// Subscribe to project 1
	if err := client.Subscribe(1); err != nil {
		t.Fatalf("Failed to subscribe to project 1: %v", err)
	}
	time.Sleep(50 * time.Millisecond) // Allow subscription to register

	// Publish event for project 1 - should receive
	event1 := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}
	if err := server.Broadcast(event1); err != nil {
		t.Fatalf("Failed to broadcast event1: %v", err)
	}

	receivedEvent := waitForEvent(t, eventChan, 1*time.Second)
	if receivedEvent.ProjectID != 1 {
		t.Errorf("Expected event for project 1, got: %d", receivedEvent.ProjectID)
	}
	t.Logf("✓ Received event for subscribed project 1")

	// Publish event for project 2 - should NOT receive (wrong project)
	event2 := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 2,
		Timestamp: time.Now(),
	}
	if err := server.Broadcast(event2); err != nil {
		t.Fatalf("Failed to broadcast event2: %v", err)
	}

	waitForNoEvent(t, eventChan, 300*time.Millisecond)
	t.Logf("✓ Correctly did not receive event for unsubscribed project 2")

	// Switch subscription to project 2
	if err := client.Subscribe(2); err != nil {
		t.Fatalf("Failed to subscribe to project 2: %v", err)
	}
	time.Sleep(50 * time.Millisecond) // Allow subscription to register

	// Publish event for project 2 - should NOW receive
	event3 := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 2,
		Timestamp: time.Now(),
	}
	if err := server.Broadcast(event3); err != nil {
		t.Fatalf("Failed to broadcast event3: %v", err)
	}

	receivedEvent = waitForEvent(t, eventChan, 1*time.Second)
	if receivedEvent.ProjectID != 2 {
		t.Errorf("Expected event for project 2, got: %d", receivedEvent.ProjectID)
	}
	t.Logf("✓ Received event after switching to project 2")

	// Publish event for project 1 - should NOT receive (switched away)
	event4 := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}
	if err := server.Broadcast(event4); err != nil {
		t.Fatalf("Failed to broadcast event4: %v", err)
	}

	waitForNoEvent(t, eventChan, 300*time.Millisecond)
	t.Logf("✓ Correctly did not receive event for previous project 1 after switch")
}

// TestSubscription_ProjectZeroHandling tests that events with ProjectID=0
// (broadcast to all) are correctly received by all clients regardless of
// their subscription.
//
// Bug context: The fix includes handling ProjectID=0 as a broadcast signal.
func TestSubscription_ProjectZeroHandling(t *testing.T) {
	server, socketPath := setupTestDaemon(t)
	logServerState(t, server, "initial")

	// Create client subscribed to project 1
	client1 := setupTestClient(t, socketPath)
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()
	eventChan1, err := client1.Listen(ctx1)
	if err != nil {
		t.Fatalf("Failed to start listening on client1: %v", err)
	}
	if err := client1.Subscribe(1); err != nil {
		t.Fatalf("Failed to subscribe client1 to project 1: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	// Create client subscribed to project 2
	client2 := setupTestClient(t, socketPath)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	eventChan2, err := client2.Listen(ctx2)
	if err != nil {
		t.Fatalf("Failed to start listening on client2: %v", err)
	}
	if err := client2.Subscribe(2); err != nil {
		t.Fatalf("Failed to subscribe client2 to project 2: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	// Publish event with ProjectID=0 (broadcast to all)
	broadcastEvent := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 0, // Broadcast
		Timestamp: time.Now(),
	}
	if err := server.Broadcast(broadcastEvent); err != nil {
		t.Fatalf("Failed to broadcast event: %v", err)
	}

	// Both clients should receive the broadcast event
	receivedEvent1 := waitForEvent(t, eventChan1, 1*time.Second)
	if receivedEvent1.ProjectID != 0 {
		t.Errorf("Client1: Expected broadcast event (ProjectID=0), got: %d", receivedEvent1.ProjectID)
	}
	t.Logf("✓ Client1 (subscribed to project 1) received broadcast event")

	receivedEvent2 := waitForEvent(t, eventChan2, 1*time.Second)
	if receivedEvent2.ProjectID != 0 {
		t.Errorf("Client2: Expected broadcast event (ProjectID=0), got: %d", receivedEvent2.ProjectID)
	}
	t.Logf("✓ Client2 (subscribed to project 2) received broadcast event")

	// Now publish a project-specific event - only matching client should receive
	project1Event := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}
	if err := server.Broadcast(project1Event); err != nil {
		t.Fatalf("Failed to broadcast project1Event: %v", err)
	}

	receivedEvent1 = waitForEvent(t, eventChan1, 1*time.Second)
	if receivedEvent1.ProjectID != 1 {
		t.Errorf("Client1: Expected project 1 event, got: %d", receivedEvent1.ProjectID)
	}
	t.Logf("✓ Client1 received project-specific event")

	waitForNoEvent(t, eventChan2, 300*time.Millisecond)
	t.Logf("✓ Client2 correctly did not receive project 1 event")
}
