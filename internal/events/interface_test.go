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

	// Test with nil concrete type
	var client *Client
	publisher = client

	// Should still be nil-checkable (interface holding nil pointer is not nil itself,
	// but we can check the concrete value)
	if publisher == nil {
		t.Error("Interface holding nil pointer should not equal nil")
	}

	// But calling methods on it would panic, so we need to check the concrete value
	// This is the pattern used in the codebase: if eventClient != nil
}

// TestEventPublisherImplementation verifies Client implements EventPublisher
func TestEventPublisherImplementation(t *testing.T) {
	// This will fail to compile if Client doesn't implement EventPublisher
	var _ EventPublisher = (*Client)(nil)
	t.Log("Client correctly implements EventPublisher interface")
}
