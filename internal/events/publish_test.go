package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockEventPublisher is a mock implementation of EventPublisher for testing
type mockRetryPublisher struct {
	sendAttempts int
	failUntil    int // Fail until this attempt number (0-indexed)
	lastEvent    Event
}

func (m *mockRetryPublisher) SendEvent(event Event) error {
	m.lastEvent = event
	currentAttempt := m.sendAttempts
	m.sendAttempts++

	if currentAttempt < m.failUntil {
		return errors.New("simulated send failure")
	}
	return nil
}

// Unused interface methods
func (m *mockRetryPublisher) Connect(ctx context.Context) error                { return nil }
func (m *mockRetryPublisher) Listen(ctx context.Context) (<-chan Event, error) { return nil, nil }
func (m *mockRetryPublisher) Subscribe(projectID int) error                    { return nil }
func (m *mockRetryPublisher) SetNotifyFunc(fn NotifyFunc)                      {}
func (m *mockRetryPublisher) Close() error                                     { return nil }

func TestPublishWithRetry_Success(t *testing.T) {
	mock := &mockRetryPublisher{failUntil: 0}
	event := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 1,
	}

	err := PublishWithRetry(mock, event, 3)
	assert.NoError(t, err)

	assert.Equal(t, 1, mock.sendAttempts)

	assert.Equal(t, 1, mock.lastEvent.ProjectID)
}

func TestPublishWithRetry_SuccessAfterRetries(t *testing.T) {
	// Fail first 2 attempts, succeed on 3rd
	mock := &mockRetryPublisher{failUntil: 2}
	event := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 2,
	}

	err := PublishWithRetry(mock, event, 3)
	assert.NoError(t, err)

	assert.Equal(t, 3, mock.sendAttempts)
}

func TestPublishWithRetry_FailureAfterAllRetries(t *testing.T) {
	// Fail all attempts
	mock := &mockRetryPublisher{failUntil: 999}
	event := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 3,
	}

	err := PublishWithRetry(mock, event, 3)
	assert.Error(t, err)

	assert.Equal(t, 3, mock.sendAttempts)

	assert.Equal(t, "simulated send failure", err.Error())
}

func TestPublishWithRetry_NilClient(t *testing.T) {
	event := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 1,
	}

	// Should not panic and return nil
	err := PublishWithRetry(nil, event, 3)
	assert.NoError(t, err)
}

func TestPublishWithRetry_ExponentialBackoff(t *testing.T) {
	// Fail first 2 attempts to trigger backoff
	mock := &mockRetryPublisher{failUntil: 2}
	event := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 4,
	}

	start := time.Now()
	err := PublishWithRetry(mock, event, 3)
	duration := time.Since(start)

	assert.NoError(t, err)

	// First retry: 50ms, Second retry: 100ms = 150ms minimum
	// Add some tolerance for test execution overhead
	minDuration := 150 * time.Millisecond
	maxDuration := 500 * time.Millisecond

	assert.GreaterOrEqual(t, duration, minDuration)

	assert.LessOrEqual(t, duration, maxDuration)
}

func TestPublishWithRetry_ZeroRetries(t *testing.T) {
	mock := &mockRetryPublisher{failUntil: 999}
	event := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 5,
	}

	// With 0 retries, should not attempt any sends
	err := PublishWithRetry(mock, event, 0)
	assert.NoError(t, err)

	assert.Equal(t, 0, mock.sendAttempts)
}
