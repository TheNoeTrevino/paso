package models

import (
	"slices"
	"time"
)

// ActivityType distinguishes between different kinds of activity items
type ActivityType int

const (
	ActivityTypeEvent   ActivityType = iota // Immutable audit event
	ActivityTypeComment                     // User comment (editable)
)

// String returns a human-readable representation of the ActivityType
func (t ActivityType) String() string {
	switch t {
	case ActivityTypeEvent:
		return "Event"
	case ActivityTypeComment:
		return "Comment"
	default:
		return "Unknown"
	}
}

// ActivityItem represents a unified view of task activity for TUI display.
// It can represent either a TaskEvent or a Comment, distinguished by the Type field.
type ActivityItem struct {
	ID        int          // Original ID from event or comment
	TaskID    int          // The task this activity belongs to
	Type      ActivityType // Distinguishes event vs comment
	Content   string       // The message/content text
	Author    string       // Who created this activity
	CreatedAt time.Time    // When the activity was created
	UpdatedAt *time.Time   // When last updated (only set for comments)
}

// NewActivityItemFromEvent creates an ActivityItem from a TaskEvent
func NewActivityItemFromEvent(event TaskEvent) ActivityItem {
	return ActivityItem{
		ID:        event.ID,
		TaskID:    event.TaskID,
		Type:      ActivityTypeEvent,
		Content:   event.Content,
		Author:    event.Author,
		CreatedAt: event.CreatedAt,
		UpdatedAt: nil,
	}
}

// NewActivityItemFromComment creates an ActivityItem from a Comment
func NewActivityItemFromComment(comment Comment) ActivityItem {
	return ActivityItem{
		ID:        comment.ID,
		TaskID:    comment.TaskID,
		Type:      ActivityTypeComment,
		Content:   comment.Message,
		Author:    comment.Author,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: &comment.UpdatedAt,
	}
}

// MergeActivities combines events and comments into a unified activity list.
// The result is sorted by CreatedAt in descending order (newest first).
func MergeActivities(events []TaskEvent, comments []Comment) []ActivityItem {
	activities := make([]ActivityItem, 0, len(events)+len(comments))

	for _, event := range events {
		activities = append(activities, NewActivityItemFromEvent(event))
	}

	for _, comment := range comments {
		activities = append(activities, NewActivityItemFromComment(comment))
	}

	slices.SortFunc(activities, func(a, b ActivityItem) int {
		return b.CreatedAt.Compare(a.CreatedAt)
	})

	return activities
}
