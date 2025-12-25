package events

import (
	"context"
	"testing"
	"time"
)

// TestNilClientMethods verifies that calling methods on a nil *Client doesn't panic
func TestNilClientMethods(t *testing.T) {
	var client *Client // nil client

	t.Run("SetNotifyFunc on nil client", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetNotifyFunc panicked on nil client: %v", r)
			}
		}()
		client.SetNotifyFunc(func(level, message string) {})
	})

	t.Run("Listen on nil client", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Listen panicked on nil client: %v", r)
			}
		}()
		ctx := context.Background()
		eventChan, err := client.Listen(ctx)
		if err == nil {
			t.Error("expected error from Listen on nil client")
		}
		// Channel should be closed
		select {
		case _, ok := <-eventChan:
			if ok {
				t.Error("expected closed channel from nil client Listen")
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("channel should be immediately readable (closed)")
		}
	})

	t.Run("Subscribe on nil client", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Subscribe panicked on nil client: %v", r)
			}
		}()
		err := client.Subscribe(1)
		if err == nil {
			t.Error("expected error from Subscribe on nil client")
		}
	})

	t.Run("SendEvent on nil client", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SendEvent panicked on nil client: %v", r)
			}
		}()
		err := client.SendEvent(Event{Type: EventDatabaseChanged})
		if err == nil {
			t.Error("expected error from SendEvent on nil client")
		}
	})

	t.Run("Close on nil client", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Close panicked on nil client: %v", r)
			}
		}()
		err := client.Close()
		if err != nil {
			t.Errorf("Close should return nil error on nil client, got: %v", err)
		}
	})
}
