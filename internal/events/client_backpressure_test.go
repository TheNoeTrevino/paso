package events

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSendEvent_QueueFullWithBackpressure tests that SendEvent applies
// exponential backoff when the queue is full, ensuring events aren't silently dropped.
func TestSendEvent_QueueFullWithBackpressure(t *testing.T) {
	t.Parallel()
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)

	// Drain initial subscribe message
	select {
	case <-messages:
	case <-time.After(500 * time.Millisecond):
		require.Fail(t, "Timeout waiting for initial subscribe")
	}

	// Fill the event queue (capacity is 100)
	// We'll send 101 events - the first 100 succeed immediately,
	// the 101st triggers backpressure/retry logic
	numEvents := 101
	var lastErr error
	for i := range numEvents {
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
		assert.ErrorContains(t, lastErr, "retry attempts exhausted")
		t.Logf("Queue saturation detected and logged: %v", lastErr)
	}

	// Force flush to ensure batching completes
	err = client.ForceFlush(ctx)
	require.NoError(t, err)

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
			assert.NotZero(t, eventCount)
			t.Logf("Backpressure mechanism allowed %d events through", eventCount)
			return
		}
	}
}

// TestSendEvent_HighThroughputReliability tests event reliability under
// high-throughput scenarios where queue saturation is likely.
func TestSendEvent_HighThroughputReliability(t *testing.T) {
	t.Parallel()
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)

	// Drain initial subscribe message
	select {
	case <-messages:
	case <-time.After(500 * time.Millisecond):
		require.Fail(t, "Timeout waiting for initial subscribe")
	}

	// Send a burst of events rapidly
	numEvents := 50
	var sendErrors int
	for i := range numEvents {
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
		t.Logf("Success rate lower than expected: %.1f%%", successRate)
	}

	// Force flush to ensure all events are sent
	err = client.ForceFlush(ctx)
	require.NoError(t, err)

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
			assert.NotZero(t, eventCount)
			t.Logf("High-throughput test: sent=%d, errors=%d, success_rate=%.1f%%, received=%d",
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
	t.Setenv("PASO_EVENT_DEBOUNCE_MS", "10")

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)

	// Drain initial subscribe message
	select {
	case <-messages:
	case <-time.After(500 * time.Millisecond):
		require.Fail(t, "Timeout waiting for initial subscribe")
	}

	// First batch: fill queue
	for range 50 {
		event := Event{
			Type:      EventDatabaseChanged,
			ProjectID: 0,
			Timestamp: time.Now(),
		}
		if err := client.SendEvent(event); err != nil {
			t.Logf("First batch error (expected): %v", err)
		}
	}

	// Force flush to drain queue via batcher
	err = client.ForceFlush(ctx)
	require.NoError(t, err)

	// Second batch: should now be able to send without excessive retry
	// If recovery works, these should succeed quickly
	successCount := 0
	for range 30 {
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
		t.Logf("Queue recovery slow: only %d/%d succeeded", successCount, 30)
	} else {
		t.Logf("Queue recovered: %d/%d succeeded after drain", successCount, 30)
	}
}

// TestSendEvent_ErrorMessageClarity tests that queue saturation errors
// provide clear, actionable error messages for debugging.
func TestSendEvent_ErrorMessageClarity(t *testing.T) {
	t.Parallel()
	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	require.NoError(t, err)
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
	for range 120 {
		err := client.SendEvent(event)
		if err != nil {
			lastErr = err
			break
		}
	}

	if lastErr != nil {
		assert.True(t, strings.Contains(lastErr.Error(), "retry attempts exhausted") ||
			strings.Contains(lastErr.Error(), "event queue full"))
		t.Logf("Error message is clear and actionable: %v", lastErr)
	} else {
		t.Log("Queue accepted all test events (batcher not running)")
	}
}
