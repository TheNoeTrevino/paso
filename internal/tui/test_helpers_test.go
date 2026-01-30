package tui

import (
	"context"
	"sync"

	"github.com/thenoetrevino/paso/internal/events"
)

// mockEventPublisher is a mock implementation of events.EventPublisher for TUI testing.
type mockEventPublisher struct {
	mu                  sync.Mutex
	subscriptionHistory []int // Track all Subscribe(projectID) calls
	currentSubscription int   // Track most recent subscription
}

func newMockEventPublisher() *mockEventPublisher {
	return &mockEventPublisher{
		subscriptionHistory: []int{},
	}
}

func (m *mockEventPublisher) Connect(ctx context.Context) error {
	return nil
}

func (m *mockEventPublisher) SendEvent(event events.Event) error {
	return nil
}

func (m *mockEventPublisher) Listen(ctx context.Context) (<-chan events.Event, error) {
	ch := make(chan events.Event)
	close(ch)
	return ch, nil
}

func (m *mockEventPublisher) Subscribe(projectID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscriptionHistory = append(m.subscriptionHistory, projectID)
	m.currentSubscription = projectID
	return nil
}

func (m *mockEventPublisher) SetNotifyFunc(fn events.NotifyFunc) {
	// No-op
}

func (m *mockEventPublisher) Close() error {
	return nil
}

func (m *mockEventPublisher) getSubscriptionHistory() []int {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]int, len(m.subscriptionHistory))
	copy(result, m.subscriptionHistory)
	return result
}

func (m *mockEventPublisher) getCurrentSubscription() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentSubscription
}
