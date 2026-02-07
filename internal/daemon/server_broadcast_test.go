package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/thenoetrevino/paso/internal/events"
)

func TestBroadcast_SingleClient(t *testing.T) {
	server, socketPath := setupTestDaemon(t)

	// Connect client using event client
	client := setupTestClient(t, socketPath)

	// Start listening for events
	eventChan, err := client.Listen(context.Background())
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	// Subscribe to project 1 and wait for subscription to register
	if err := client.Subscribe(1); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	waitForSubscriptions(t, server, 1)

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

		eventChan, err := client.Listen(context.Background())
		if err != nil {
			t.Fatalf("Client %d failed to listen: %v", i, err)
		}
		eventChans = append(eventChans, eventChan)

		// Subscribe all to project 1
		if err := client.Subscribe(1); err != nil {
			t.Fatalf("Client %d failed to subscribe: %v", i, err)
		}
	}

	// Wait for all subscriptions to register
	waitForSubscriptions(t, server, numClients)

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
	eventChanA, _ := clientA.Listen(context.Background())
	if err := clientA.Subscribe(1); err != nil {
		t.Fatalf("ClientA failed to subscribe: %v", err)
	}
	waitForSubscriptions(t, server, 1)

	// Client B subscribes to project 2
	clientB := setupTestClient(t, socketPath)
	eventChanB, _ := clientB.Listen(context.Background())
	if err := clientB.Subscribe(2); err != nil {
		t.Fatalf("ClientB failed to subscribe: %v", err)
	}
	waitForSubscriptions(t, server, 1)

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
	eventChanA, _ := clientA.Listen(context.Background())
	if err := clientA.Subscribe(1); err != nil {
		t.Fatalf("ClientA failed to subscribe: %v", err)
	}
	waitForSubscriptions(t, server, 1)

	// Client B subscribes to project 2
	clientB := setupTestClient(t, socketPath)
	eventChanB, _ := clientB.Listen(context.Background())
	if err := clientB.Subscribe(2); err != nil {
		t.Fatalf("ClientB failed to subscribe: %v", err)
	}
	waitForSubscriptions(t, server, 1)

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
	eventChan, _ := client.Listen(context.Background())
	if err := client.Subscribe(1); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	waitForSubscriptions(t, server, 1)

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
