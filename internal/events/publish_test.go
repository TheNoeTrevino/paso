package events

import (
	"context"
	"errors"
	"testing"
	"time"
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
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if mock.sendAttempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", mock.sendAttempts)
	}

	if mock.lastEvent.ProjectID != 1 {
		t.Errorf("Expected event project ID 1, got %d", mock.lastEvent.ProjectID)
	}
}

func TestPublishWithRetry_SuccessAfterRetries(t *testing.T) {
	// Fail first 2 attempts, succeed on 3rd
	mock := &mockRetryPublisher{failUntil: 2}
	event := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 2,
	}

	err := PublishWithRetry(mock, event, 3)
	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}

	if mock.sendAttempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", mock.sendAttempts)
	}
}

func TestPublishWithRetry_FailureAfterAllRetries(t *testing.T) {
	// Fail all attempts
	mock := &mockRetryPublisher{failUntil: 999}
	event := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 3,
	}

	err := PublishWithRetry(mock, event, 3)
	if err == nil {
		t.Error("Expected error after all retries failed, got nil")
	}

	if mock.sendAttempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", mock.sendAttempts)
	}

	expectedErr := "simulated send failure"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestPublishWithRetry_NilClient(t *testing.T) {
	event := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 1,
	}

	// Should not panic and return nil
	err := PublishWithRetry(nil, event, 3)
	if err != nil {
		t.Errorf("Expected nil error for nil client, got: %v", err)
	}
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

	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}

	// First retry: 50ms, Second retry: 100ms = 150ms minimum
	// Add some tolerance for test execution overhead
	minDuration := 150 * time.Millisecond
	maxDuration := 500 * time.Millisecond

	if duration < minDuration {
		t.Errorf("Expected at least %v delay for retries, got %v", minDuration, duration)
	}

	if duration > maxDuration {
		t.Errorf("Expected delay under %v, got %v (may indicate backoff is too long)", maxDuration, duration)
	}
}

func TestPublishWithRetry_ZeroRetries(t *testing.T) {
	mock := &mockRetryPublisher{failUntil: 999}
	event := Event{
		Type:      EventDatabaseChanged,
		ProjectID: 5,
	}

	// With 0 retries, should not attempt any sends
	err := PublishWithRetry(mock, event, 0)
	if err != nil {
		t.Errorf("Expected nil error with 0 retries (no attempts), got: %v", err)
	}

	if mock.sendAttempts != 0 {
		t.Errorf("Expected 0 attempts with maxRetries=0, got %d", mock.sendAttempts)
	}
}
