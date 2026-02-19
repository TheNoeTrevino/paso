package state

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd_BasicNotification(t *testing.T) {
	t.Parallel()
	ns := NewNotificationState()

	id := ns.Add(LevelInfo, "Hello")

	assert.NotEqual(t, 0, id, "Add should return a non-zero ID")
	require.Len(t, ns.All(), 1)
	assert.Equal(t, "Hello", ns.All()[0].Message)
	assert.Equal(t, LevelInfo, ns.All()[0].Level)
}

func TestAdd_DebounceSameMessage(t *testing.T) {
	t.Parallel()
	ns := NewNotificationState()

	id1 := ns.Add(LevelInfo, "Already at the last column")
	id2 := ns.Add(LevelInfo, "Already at the last column")

	assert.NotEqual(t, 0, id1, "first add should succeed")
	assert.Equal(t, 0, id2, "duplicate within debounce window should be suppressed")
	assert.Len(t, ns.All(), 1, "only one notification should exist")
}

func TestAdd_DifferentMessagesNotDebounced(t *testing.T) {
	t.Parallel()
	ns := NewNotificationState()

	id1 := ns.Add(LevelInfo, "Already at the last column")
	id2 := ns.Add(LevelInfo, "Already at the first task")

	assert.NotEqual(t, 0, id1)
	assert.NotEqual(t, 0, id2)
	assert.Len(t, ns.All(), 2, "different messages should both be added")
}

func TestAdd_DifferentLevelsSameMessage(t *testing.T) {
	t.Parallel()
	ns := NewNotificationState()

	id1 := ns.Add(LevelInfo, "No task selected")
	id2 := ns.Add(LevelError, "No task selected")

	assert.NotEqual(t, 0, id1)
	assert.NotEqual(t, 0, id2)
	assert.Len(t, ns.All(), 2, "same message at different levels should both be added")
}

func TestAdd_DebounceExpiresAfterDuration(t *testing.T) {
	t.Parallel()
	ns := NewNotificationState()
	ns.debounceDuration = 1 * time.Millisecond

	id1 := ns.Add(LevelInfo, "Already at the last column")
	assert.NotEqual(t, 0, id1)

	time.Sleep(2 * time.Millisecond)

	id2 := ns.Add(LevelInfo, "Already at the last column")
	assert.NotEqual(t, 0, id2, "should succeed after debounce window expires")
	assert.Len(t, ns.All(), 2)
}

func TestAdd_MaxNotificationsEvictionStillWorks(t *testing.T) {
	t.Parallel()
	ns := NewNotificationState()

	ns.Add(LevelInfo, "msg 1")
	ns.Add(LevelInfo, "msg 2")
	ns.Add(LevelInfo, "msg 3")
	ns.Add(LevelInfo, "msg 4")

	assert.Len(t, ns.All(), maxNotifications, "should evict oldest when exceeding max")
	assert.Equal(t, "msg 2", ns.All()[0].Message, "oldest should be evicted")
	assert.Equal(t, "msg 4", ns.All()[2].Message, "newest should be last")
}

func TestRemove_ZeroIDNoOp(t *testing.T) {
	t.Parallel()
	ns := NewNotificationState()

	ns.Add(LevelInfo, "keep me")

	// Remove with ID 0 (suppressed sentinel) should not panic or remove anything
	ns.Remove(0)

	assert.Len(t, ns.All(), 1, "Remove(0) should be a no-op")
	assert.Equal(t, "keep me", ns.All()[0].Message)
}
