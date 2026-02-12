package events_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
)

func TestPublishWithRetry_Success(t *testing.T) {
	t.Parallel()
	mock := mocks.NewMockEventPublisher()
	sendAttempts := 0
	mock.SendEventFunc = func(event events.Event) error {
		sendAttempts++
		return nil
	}

	event := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 1,
	}

	err := events.PublishWithRetry(mock, event, 3)
	assert.NoError(t, err)
	assert.Equal(t, 1, sendAttempts)
	assert.Equal(t, 1, mock.SentEvents[len(mock.SentEvents)-1].ProjectID)
}

func TestPublishWithRetry_SuccessAfterRetries(t *testing.T) {
	t.Parallel()
	mock := mocks.NewMockEventPublisher()
	sendAttempts := 0
	failUntil := 2
	mock.SendEventFunc = func(event events.Event) error {
		attempt := sendAttempts
		sendAttempts++
		if attempt < failUntil {
			return errors.New("simulated send failure")
		}
		return nil
	}

	event := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 2,
	}

	err := events.PublishWithRetry(mock, event, 3)
	assert.NoError(t, err)
	assert.Equal(t, 3, sendAttempts)
}

func TestPublishWithRetry_FailureAfterAllRetries(t *testing.T) {
	t.Parallel()
	mock := mocks.NewMockEventPublisher()
	sendAttempts := 0
	failUntil := 999
	mock.SendEventFunc = func(event events.Event) error {
		attempt := sendAttempts
		sendAttempts++
		if attempt < failUntil {
			return errors.New("simulated send failure")
		}
		return nil
	}

	event := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 3,
	}

	err := events.PublishWithRetry(mock, event, 3)
	assert.Error(t, err)
	assert.Equal(t, 3, sendAttempts)
	assert.ErrorContains(t, err, "simulated send failure")
}

func TestPublishWithRetry_NilClient(t *testing.T) {
	t.Parallel()
	event := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 1,
	}

	err := events.PublishWithRetry(nil, event, 3)
	assert.NoError(t, err)
}

func TestPublishWithRetry_ExponentialBackoff(t *testing.T) {
	t.Parallel()
	mock := mocks.NewMockEventPublisher()
	sendAttempts := 0
	failUntil := 2
	mock.SendEventFunc = func(event events.Event) error {
		attempt := sendAttempts
		sendAttempts++
		if attempt < failUntil {
			return errors.New("simulated send failure")
		}
		return nil
	}

	event := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 4,
	}

	start := time.Now()
	err := events.PublishWithRetry(mock, event, 3)
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
	t.Parallel()
	mock := mocks.NewMockEventPublisher()
	sendAttempts := 0
	mock.SendEventFunc = func(event events.Event) error {
		sendAttempts++
		return errors.New("simulated send failure")
	}

	event := events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: 5,
	}

	err := events.PublishWithRetry(mock, event, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, sendAttempts)
}
