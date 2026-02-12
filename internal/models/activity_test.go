package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivityType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    ActivityType
		expected string
	}{
		{
			name:     "event type",
			input:    ActivityTypeEvent,
			expected: "Event",
		},
		{
			name:     "comment type",
			input:    ActivityTypeComment,
			expected: "Comment",
		},
		{
			name:     "unknown type",
			input:    ActivityType(99),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()

			result := tt.input.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewActivityItemFromEvent(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	event := TaskEvent{
		ID:        1,
		TaskID:    10,
		Content:   "Task moved to In Progress",
		Author:    "alice",
		CreatedAt: now,
	}

	item := NewActivityItemFromEvent(event)

	assert.Equal(t, 1, item.ID)
	assert.Equal(t, 10, item.TaskID)
	assert.Equal(t, ActivityTypeEvent, item.Type)
	assert.Equal(t, "Task moved to In Progress", item.Content)
	assert.Equal(t, "alice", item.Author)
	assert.Equal(t, now, item.CreatedAt)
	assert.Nil(t, item.UpdatedAt, "events should not have UpdatedAt")
}

func TestNewActivityItemFromComment(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	updated := time.Date(2025, 6, 15, 11, 0, 0, 0, time.UTC)
	comment := Comment{
		ID:        5,
		TaskID:    10,
		Message:   "Looking good",
		Author:    "bob",
		CreatedAt: now,
		UpdatedAt: updated,
	}

	item := NewActivityItemFromComment(comment)

	assert.Equal(t, 5, item.ID)
	assert.Equal(t, 10, item.TaskID)
	assert.Equal(t, ActivityTypeComment, item.Type)
	assert.Equal(t, "Looking good", item.Content)
	assert.Equal(t, "bob", item.Author)
	assert.Equal(t, now, item.CreatedAt)
	require.NotNil(t, item.UpdatedAt, "comments should have UpdatedAt")
	assert.Equal(t, updated, *item.UpdatedAt)
}

func TestMergeActivities(t *testing.T) {
	t.Parallel()

	t1 := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 6, 15, 11, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	t4 := time.Date(2025, 6, 15, 13, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		events           []TaskEvent
		comments         []Comment
		expectedLen      int
		expectedFirstID  int
		expectedFirstTyp ActivityType
		expectedLastID   int
		expectedLastTyp  ActivityType
	}{
		{
			name:        "both empty",
			events:      nil,
			comments:    nil,
			expectedLen: 0,
		},
		{
			name:        "empty events and comments slices",
			events:      []TaskEvent{},
			comments:    []Comment{},
			expectedLen: 0,
		},
		{
			name: "events only sorted newest first",
			events: []TaskEvent{
				{ID: 1, TaskID: 10, Content: "created", Author: "a", CreatedAt: t1},
				{ID: 2, TaskID: 10, Content: "moved", Author: "a", CreatedAt: t3},
			},
			comments:         nil,
			expectedLen:      2,
			expectedFirstID:  2,
			expectedFirstTyp: ActivityTypeEvent,
			expectedLastID:   1,
			expectedLastTyp:  ActivityTypeEvent,
		},
		{
			name:   "comments only sorted newest first",
			events: nil,
			comments: []Comment{
				{ID: 5, TaskID: 10, Message: "first", Author: "b", CreatedAt: t1, UpdatedAt: t1},
				{ID: 6, TaskID: 10, Message: "second", Author: "b", CreatedAt: t2, UpdatedAt: t2},
			},
			expectedLen:      2,
			expectedFirstID:  6,
			expectedFirstTyp: ActivityTypeComment,
			expectedLastID:   5,
			expectedLastTyp:  ActivityTypeComment,
		},
		{
			name: "mixed events and comments interleaved by time",
			events: []TaskEvent{
				{ID: 1, TaskID: 10, Content: "created", Author: "a", CreatedAt: t1},
				{ID: 2, TaskID: 10, Content: "moved", Author: "a", CreatedAt: t3},
			},
			comments: []Comment{
				{ID: 5, TaskID: 10, Message: "note", Author: "b", CreatedAt: t2, UpdatedAt: t2},
				{ID: 6, TaskID: 10, Message: "latest", Author: "b", CreatedAt: t4, UpdatedAt: t4},
			},
			expectedLen:      4,
			expectedFirstID:  6,
			expectedFirstTyp: ActivityTypeComment,
			expectedLastID:   1,
			expectedLastTyp:  ActivityTypeEvent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()

			result := MergeActivities(tt.events, tt.comments)
			require.Len(t, result, tt.expectedLen)

			if tt.expectedLen == 0 {
				return
			}

			assert.Equal(t, tt.expectedFirstID, result[0].ID, "first item ID (newest)")
			assert.Equal(t, tt.expectedFirstTyp, result[0].Type, "first item type")
			assert.Equal(t, tt.expectedLastID, result[tt.expectedLen-1].ID, "last item ID (oldest)")
			assert.Equal(t, tt.expectedLastTyp, result[tt.expectedLen-1].Type, "last item type")

			// Verify descending order
			for i := 1; i < len(result); i++ {
				assert.False(t, result[i].CreatedAt.After(result[i-1].CreatedAt),
					"activities should be sorted newest first: item %d (%v) should not be after item %d (%v)",
					i, result[i].CreatedAt, i-1, result[i-1].CreatedAt)
			}
		})
	}
}
