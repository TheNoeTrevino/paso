package mocks

import (
	"context"
	"sync"

	"github.com/thenoetrevino/paso/internal/events"
)

// Compile-time interface verification
var _ events.EventPublisher = (*MockEventPublisher)(nil)

// MockEventPublisher is a mock implementation of events.EventPublisher for testing.
// It records all method calls and published events for verification in tests.
type MockEventPublisher struct {
	mu    sync.Mutex
	Calls []MockCall

	// Per-method error injection
	ConnectErr    error
	SendEventErr  error
	ListenErr     error
	SubscribeErr  error
	ForceFlushErr error
	CloseErr      error

	// Per-method callback injection
	SendEventFunc func(events.Event) error

	// Result injection
	ListenResult <-chan events.Event

	// Event tracking
	SentEvents []events.Event

	// Call tracking
	CloseCalled     bool
	ConnectCalled   bool
	SubscribeCalled bool
	ListenCalled    bool

	// Subscription tracking
	SubscriptionHistory []int
	CurrentSubscription int
}

// NewMockEventPublisher creates a new mock event publisher.
func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{
		Calls:               make([]MockCall, 0),
		SentEvents:          []events.Event{},
		SubscriptionHistory: []int{},
	}
}

func (m *MockEventPublisher) recordCall(method string, args map[string]any) {
	m.Calls = append(m.Calls, MockCall{
		Method: method,
		Args:   args,
	})
}

// Reset clears all recorded calls, events, and tracking state.
func (m *MockEventPublisher) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = make([]MockCall, 0)
	m.SentEvents = []events.Event{}
	m.CloseCalled = false
	m.ConnectCalled = false
	m.SubscribeCalled = false
	m.ListenCalled = false
	m.SubscriptionHistory = []int{}
	m.CurrentSubscription = 0
	m.SendEventFunc = nil
}

// GetCalls returns a copy of all recorded calls.
func (m *MockEventPublisher) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockCall, len(m.Calls))
	copy(result, m.Calls)
	return result
}

// HasCall checks if a method was called.
func (m *MockEventPublisher) HasCall(method string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.Calls {
		if call.Method == method {
			return true
		}
	}
	return false
}

// CallCount returns the number of times a method was called.
func (m *MockEventPublisher) CallCount(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, call := range m.Calls {
		if call.Method == method {
			count++
		}
	}
	return count
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

func (m *MockEventPublisher) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ConnectCalled = true
	m.recordCall("Connect", nil)
	return m.ConnectErr
}

func (m *MockEventPublisher) SendEvent(event events.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SentEvents = append(m.SentEvents, event)
	m.recordCall("SendEvent", map[string]any{
		"event": event,
	})
	if m.SendEventFunc != nil {
		return m.SendEventFunc(event)
	}
	return m.SendEventErr
}

func (m *MockEventPublisher) Listen(ctx context.Context) (<-chan events.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ListenCalled = true
	m.recordCall("Listen", nil)
	if m.ListenErr != nil {
		return nil, m.ListenErr
	}
	if m.ListenResult != nil {
		return m.ListenResult, nil
	}
	ch := make(chan events.Event)
	close(ch)
	return ch, nil
}

func (m *MockEventPublisher) Subscribe(projectID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SubscribeCalled = true
	m.SubscriptionHistory = append(m.SubscriptionHistory, projectID)
	m.CurrentSubscription = projectID
	m.recordCall("Subscribe", map[string]any{
		"projectID": projectID,
	})
	return m.SubscribeErr
}

func (m *MockEventPublisher) SetNotifyFunc(fn events.NotifyFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("SetNotifyFunc", nil)
}

func (m *MockEventPublisher) ForceFlush(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("ForceFlush", nil)
	return m.ForceFlushErr
}

func (m *MockEventPublisher) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CloseCalled = true
	m.recordCall("Close", nil)
	return m.CloseErr
}
