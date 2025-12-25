package events

import (
	"testing"
)

// TestEventPublisherNilCheck verifies that nil checks work correctly with the interface
func TestEventPublisherNilCheck(t *testing.T) {
	var publisher EventPublisher

	// Should be nil
	if publisher != nil {
		t.Error("Expected nil EventPublisher to be nil")
	}

	// Test with nil concrete type - interface holding nil pointer is not nil itself
	// Methods should handle nil receiver
	var client *Client
	if client == nil {
		t.Log("Nil pointer wrapped in interface - methods should handle nil receiver")
	}

	// Verify we can assign nil client to interface (compile-time check)
	publisher = client
	_ = publisher // Use it to avoid ineffassign
}

// TestEventPublisherImplementation verifies Client implements EventPublisher
func TestEventPublisherImplementation(t *testing.T) {
	// This will fail to compile if Client doesn't implement EventPublisher
	var _ EventPublisher = (*Client)(nil)
	t.Log("Client correctly implements EventPublisher interface")
}
