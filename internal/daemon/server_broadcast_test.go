package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/events"
)

func TestBroadcast_SingleClient(t *testing.T) {
	t.Parallel()
	server, socketPath := setupTestDaemon(t)

	// Connect client using event client
	client := setupTestClient(t, socketPath)

	// Start listening for events
	eventChan, err := client.Listen(context.Background())
	require.NoError(t, err)

	// Subscribe to project 1 and wait for subscription to register
	waiter := expectSubscriptions(t, server, 1)
	err = client.Subscribe(1)
	require.NoError(t, err)
	waiter.Wait()

	// Broadcast an event
	testEvent := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}

	err = server.Broadcast(testEvent)
	require.NoError(t, err)

	// Wait for event
	receivedEvent := waitForEvent(t, eventChan, 2*time.Second)

	assert.Equal(t, 1, receivedEvent.ProjectID)

	// Verify sequence ID was set
	assert.NotZero(t, receivedEvent.SequenceID)

	t.Logf("Event broadcast and received successfully (sequence: %d)", receivedEvent.SequenceID)
}

func TestBroadcast_MultipleClients(t *testing.T) {
	t.Parallel()
	server, socketPath := setupTestDaemon(t)

	numClients := 3
	var eventChans []<-chan events.Event

	// Connect multiple clients
	waiter := expectSubscriptions(t, server, numClients)
	for i := 0; i < numClients; i++ {
		client := setupTestClient(t, socketPath)

		eventChan, err := client.Listen(context.Background())
		require.NoError(t, err)
		eventChans = append(eventChans, eventChan)

		// Subscribe all to project 1
		err = client.Subscribe(1)
		require.NoError(t, err)
	}

	// Wait for all subscriptions to register
	waiter.Wait()

	// Broadcast event
	testEvent := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}

	err := server.Broadcast(testEvent)
	require.NoError(t, err)

	// Verify all clients receive the event
	for i, eventChan := range eventChans {
		receivedEvent := waitForEvent(t, eventChan, 2*time.Second)
		assert.Equal(t, 1, receivedEvent.ProjectID)
		t.Logf("Client %d received event (sequence: %d)", i, receivedEvent.SequenceID)
	}
}

func TestBroadcast_SubscriptionFiltering(t *testing.T) {
	t.Parallel()
	server, socketPath := setupTestDaemon(t)

	// Client A subscribes to project 1
	clientA := setupTestClient(t, socketPath)
	eventChanA, _ := clientA.Listen(context.Background())
	waiterA := expectSubscriptions(t, server, 1)
	err := clientA.Subscribe(1)
	require.NoError(t, err)
	waiterA.Wait()

	// Client B subscribes to project 2
	clientB := setupTestClient(t, socketPath)
	eventChanB, _ := clientB.Listen(context.Background())
	waiterB := expectSubscriptions(t, server, 1)
	err = clientB.Subscribe(2)
	require.NoError(t, err)
	waiterB.Wait()

	// Broadcast event for project 1
	testEvent := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 1,
		Timestamp: time.Now(),
	}

	err = server.Broadcast(testEvent)
	require.NoError(t, err)

	// Client A should receive it
	receivedEvent := waitForEvent(t, eventChanA, 2*time.Second)
	assert.Equal(t, 1, receivedEvent.ProjectID)

	// Client B should NOT receive it (different project)
	waitForNoEvent(t, eventChanB, 500*time.Millisecond)

	t.Logf("Subscription filtering works correctly")
}

func TestBroadcast_AllProjects(t *testing.T) {
	t.Parallel()
	server, socketPath := setupTestDaemon(t)

	// Client A subscribes to project 1
	clientA := setupTestClient(t, socketPath)
	eventChanA, _ := clientA.Listen(context.Background())
	waiterA := expectSubscriptions(t, server, 1)
	err := clientA.Subscribe(1)
	require.NoError(t, err)
	waiterA.Wait()

	// Client B subscribes to project 2
	clientB := setupTestClient(t, socketPath)
	eventChanB, _ := clientB.Listen(context.Background())
	waiterB := expectSubscriptions(t, server, 1)
	err = clientB.Subscribe(2)
	require.NoError(t, err)
	waiterB.Wait()

	// Broadcast event for all projects (projectID = 0)
	testEvent := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 0,
		Timestamp: time.Now(),
	}

	err = server.Broadcast(testEvent)
	require.NoError(t, err)

	// Both clients should receive it
	receivedEventA := waitForEvent(t, eventChanA, 2*time.Second)
	assert.Equal(t, 0, receivedEventA.ProjectID)

	receivedEventB := waitForEvent(t, eventChanB, 2*time.Second)
	assert.Equal(t, 0, receivedEventB.ProjectID)

	t.Logf("Broadcast to all projects works correctly")
}

func TestBroadcast_SequenceNumbers(t *testing.T) {
	t.Parallel()
	server, socketPath := setupTestDaemon(t)

	client := setupTestClient(t, socketPath)
	eventChan, _ := client.Listen(context.Background())
	waiter := expectSubscriptions(t, server, 1)
	err := client.Subscribe(1)
	require.NoError(t, err)

	waiter.Wait()

	// Send 10 events
	numEvents := 10
	for i := 0; i < numEvents; i++ {
		testEvent := events.Event{
			Type:      events.EventDatabaseChanged,
			ProjectID: 1,
			Timestamp: time.Now(),
		}
		err := server.Broadcast(testEvent)
		require.NoError(t, err)
	}

	// Collect all events
	var sequences []int64
	for i := 0; i < numEvents; i++ {
		event := waitForEvent(t, eventChan, 2*time.Second)
		sequences = append(sequences, event.SequenceID)
	}

	// Verify sequences are monotonically increasing
	for i := 1; i < len(sequences); i++ {
		assert.Less(t, sequences[i-1], sequences[i])
	}

	t.Logf("Sequence numbers are monotonically increasing: %v", sequences)
}
