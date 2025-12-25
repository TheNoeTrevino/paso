package events_test

import (
	"context"
	"sync"

	"github.com/thenoetrevino/paso/internal/events"
)

// MockEventPublisher is a mock implementation of events.EventPublisher for testing.
// It records all published events for verification in tests.
type MockEventPublisher struct {
	mu sync.Mutex

	// Recorded events
	SentEvents []events.Event

	// Tracking
	CloseCalled     bool
	ConnectCalled   bool
	SubscribeCalled bool
	ListenCalled    bool

	// Subscription tracking
	SubscriptionHistory []int // Track all Subscribe(projectID) calls in order
	CurrentSubscription int   // Track the most recent subscription
}

// NewMockEventPublisher creates a new mock event publisher.
func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{
		SentEvents:          []events.Event{},
		SubscriptionHistory: []int{},
	}
}

// Connect is a no-op for the mock.
func (m *MockEventPublisher) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ConnectCalled = true
	return nil
}

// SendEvent records the event for later verification.
func (m *MockEventPublisher) SendEvent(event events.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SentEvents = append(m.SentEvents, event)
	return nil
}

// Listen is a no-op for the mock (returns empty channel).
func (m *MockEventPublisher) Listen(ctx context.Context) (<-chan events.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ListenCalled = true
	ch := make(chan events.Event)
	close(ch) // Return closed channel
	return ch, nil
}

// Subscribe is a no-op for the mock.
func (m *MockEventPublisher) Subscribe(projectID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SubscribeCalled = true
	m.SubscriptionHistory = append(m.SubscriptionHistory, projectID)
	m.CurrentSubscription = projectID
	return nil
}

// SetNotifyFunc is a no-op for the mock.
func (m *MockEventPublisher) SetNotifyFunc(fn events.NotifyFunc) {
	// No-op
}

// Close marks the publisher as closed.
func (m *MockEventPublisher) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CloseCalled = true
	return nil
}

// Reset clears all recorded events. Useful for tests that need multiple assertions.
func (m *MockEventPublisher) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SentEvents = []events.Event{}
	m.CloseCalled = false
	m.ConnectCalled = false
	m.SubscribeCalled = false
	m.ListenCalled = false
	m.SubscriptionHistory = []int{}
	m.CurrentSubscription = 0
}

// GetEventsByType returns all events of a specific type.
func (m *MockEventPublisher) GetEventsByType(eventType events.EventType) []events.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []events.Event
	for _, e := range m.SentEvents {
		if e.Type == eventType {
			result = append(result, e)
		}
	}
	return result
}

// GetEventsByProject returns all events for a specific project.
func (m *MockEventPublisher) GetEventsByProject(projectID int) []events.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []events.Event
	for _, e := range m.SentEvents {
		if e.ProjectID == projectID {
			result = append(result, e)
		}
	}
	return result
}

// AssertEventSent checks if an event with the given project ID was sent.
func (m *MockEventPublisher) AssertEventSent(projectID int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.SentEvents {
		if e.ProjectID == projectID {
			return true
		}
	}
	return false
}

// EventCount returns the total number of events sent.
func (m *MockEventPublisher) EventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.SentEvents)
}

// GetSubscriptionHistory returns all project IDs that were subscribed to.
func (m *MockEventPublisher) GetSubscriptionHistory() []int {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]int, len(m.SubscriptionHistory))
	copy(result, m.SubscriptionHistory)
	return result
}

// GetCurrentSubscription returns the most recent project ID subscribed to.
func (m *MockEventPublisher) GetCurrentSubscription() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.CurrentSubscription
}

// Compile-time interface verification
var _ events.EventPublisher = (*MockEventPublisher)(nil)
