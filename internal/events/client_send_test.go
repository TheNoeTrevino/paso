package events

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendEvent_Success(t *testing.T) {
	t.Parallel()
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)

	// Drain initial subscribe message
	select {
	case <-messages:
	case <-time.After(500 * time.Millisecond):
		require.Fail(t, "Timeout waiting for initial subscribe")
	}

	// Send an event
	testEvent := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}

	err = client.SendEvent(testEvent)
	require.NoError(t, err)

	// Force flush to ensure event is sent immediately
	err = client.ForceFlush(ctx)
	require.NoError(t, err)

	// Check if message was received (might be batched)
	select {
	case msg := <-messages:
		assert.Equal(t, "event", msg.Type)
		require.NotNil(t, msg.Event)
		t.Logf("Event sent successfully: %+v", msg.Event)
	case <-time.After(2 * time.Second):
		require.Fail(t, "Timeout waiting for event message")
	}
}

func TestSendEvent_BeforeConnect(t *testing.T) {
	t.Parallel()
	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	// Send event before connecting - should succeed (queued)
	testEvent := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 1,
	}

	err = client.SendEvent(testEvent)
	assert.NoError(t, err)

	t.Logf("SendEvent queues events before connection")
}

func TestSendEvent_Batching(t *testing.T) {
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	// Set short debounce for testing
	t.Setenv("PASO_EVENT_DEBOUNCE_MS", "50")

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)

	// Drain initial subscribe message
	select {
	case <-messages:
	case <-time.After(500 * time.Millisecond):
		require.Fail(t, "Timeout waiting for initial subscribe")
	}

	// Send multiple events rapidly
	numEvents := 5
	for i := range numEvents {
		testEvent := Event{
			Type:      EventDatabaseChanged,
			ProjectID: i,
		}
		err := client.SendEvent(testEvent)
		require.NoError(t, err)
	}

	// Force flush to ensure batch is sent
	err = client.ForceFlush(ctx)
	require.NoError(t, err)

	// Should receive at least one message (events might be batched)
	select {
	case msg := <-messages:
		assert.Equal(t, "event", msg.Type)
		t.Logf("Batched events sent successfully")
	case <-time.After(1 * time.Second):
		require.Fail(t, "Timeout waiting for batched events")
	}
}
