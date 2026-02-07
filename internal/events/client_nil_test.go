package events

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNilClientMethods verifies that calling methods on a nil *Client doesn't panic
func TestNilClientMethods(t *testing.T) {
	var client *Client // nil client

	t.Run("SetNotifyFunc on nil client", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Nil(t, r)
		}()
		client.SetNotifyFunc(func(level, message string) {})
	})

	t.Run("Listen on nil client", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Nil(t, r)
		}()
		ctx := context.Background()
		eventChan, err := client.Listen(ctx)
		require.Error(t, err)
		// Channel should be closed
		select {
		case _, ok := <-eventChan:
			assert.False(t, ok)
		case <-time.After(100 * time.Millisecond):
			assert.Fail(t, "channel should be immediately readable (closed)")
		}
	})

	t.Run("Subscribe on nil client", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Nil(t, r)
		}()
		err := client.Subscribe(1)
		assert.Error(t, err)
	})

	t.Run("SendEvent on nil client", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Nil(t, r)
		}()
		err := client.SendEvent(Event{Type: EventDatabaseChanged})
		assert.Error(t, err)
	})

	t.Run("Close on nil client", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Nil(t, r)
		}()
		err := client.Close()
		assert.NoError(t, err)
	})
}
