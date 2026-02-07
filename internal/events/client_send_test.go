package events

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

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
	t.Setenv("PASO_EVENT_DEBOUNCE_MS", "50")

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
