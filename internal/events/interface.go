package events

import "context"

// EventPublisher defines the interface for sending and receiving events.
// This interface allows for loose coupling and easier testing by depending
// on behavior rather than concrete implementation.
type EventPublisher interface {
	// Connect establishes a connection to the daemon socket
	Connect(ctx context.Context) error

	// SendEvent queues an event to be sent to the daemon
	SendEvent(event Event) error

	// Listen starts listening for events from the daemon
	Listen(ctx context.Context) (<-chan Event, error)

	// Subscribe changes the subscription to a specific project
	Subscribe(projectID int) error

	// Close closes the connection to the daemon and stops all goroutines
	Close() error
}

// Compile-time verification that *Client implements EventPublisher
var _ EventPublisher = (*Client)(nil)
