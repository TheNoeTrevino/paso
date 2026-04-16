package state

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
)

// NotificationLevel represents the severity/type of a notification.
type NotificationLevel int

const (
	// LevelInfo represents informational notifications (blue, bell icon)
	LevelInfo NotificationLevel = iota
	// LevelWarning represents warning notifications (yellow, warning icon)
	LevelWarning // currently not used
	// LevelError represents error notifications (red, error icon)
	LevelError
)

const maxNotifications = 3

// DefaultDebounceDuration is the default time window during which duplicate
// notifications (same level + message) are suppressed.
const DefaultDebounceDuration = 500 * time.Millisecond

// Notification represents a single notification message with a severity level.
type Notification struct {
	ID        int
	Level     NotificationLevel
	Message   string
	CreatedAt time.Time
}

// NotificationState manages notification display state.
// This provides a centralized way to handle user-facing notifications
// of different severity levels throughout the application.
//
// It includes per-message debouncing: if the same level+message is added
// within the debounce window, the duplicate is suppressed (Add returns 0).
type NotificationState struct {
	// notifications contains the list of current notifications to display
	notifications []Notification
	// nextID is the auto-incrementing ID counter for notifications
	nextID int
	// windowWidth tracks the current window width for positioning
	windowWidth int
	// windowHeight tracks the current window height for positioning
	windowHeight int
	// lastSeen tracks when each notification key (level:message) was last added,
	// used for per-message debouncing to suppress rapid duplicate notifications.
	lastSeen map[string]time.Time
	// debounceDuration is the time window during which duplicate notifications are suppressed.
	debounceDuration time.Duration
}

// NewNotificationState creates a new NotificationState with no notifications.
func NewNotificationState() *NotificationState {
	return &NotificationState{
		notifications:    []Notification{},
		nextID:           1,
		windowWidth:      0,
		windowHeight:     0,
		lastSeen:         make(map[string]time.Time),
		debounceDuration: DefaultDebounceDuration,
	}
}

// Add adds a new notification with the specified level and message.
// Returns the ID assigned to the new notification, or 0 if the notification
// was suppressed by debouncing (same level+message added within the debounce window).
//
// Parameters:
//   - level: the severity level of the notification
//   - message: the notification message to display
func (s *NotificationState) Add(level NotificationLevel, message string) int {
	now := time.Now()
	key := fmt.Sprintf("%d:%s", level, message)

	// Debounce: suppress if the same notification was added recently
	if lastTime, ok := s.lastSeen[key]; ok {
		if now.Sub(lastTime) < s.debounceDuration {
			return 0
		}
	}

	s.lastSeen[key] = now

	id := s.nextID
	s.notifications = append(s.notifications, Notification{
		ID:        id,
		Level:     level,
		Message:   message,
		CreatedAt: now,
	})
	s.nextID++

	// Evict oldest if we exceed max stack size
	if len(s.notifications) > maxNotifications {
		s.notifications = s.notifications[len(s.notifications)-maxNotifications:]
	}

	return id
}

// Remove removes a notification by its ID. Used for auto-dismiss expiration.
// ID 0 is a no-op (0 is the sentinel for debounce-suppressed notifications).
func (s *NotificationState) Remove(id int) {
	if id == 0 {
		return
	}
	for i, n := range s.notifications {
		if n.ID == id {
			s.notifications = append(s.notifications[:i], s.notifications[i+1:]...)
			return
		}
	}
}

// Clear removes all notifications.
func (s *NotificationState) Clear() {
	s.notifications = []Notification{}
}

// ClearLevel removes all notifications of a specific level.
//
// Parameters:
//   - level: the level of notifications to clear
func (s *NotificationState) ClearLevel(level NotificationLevel) {
	filtered := []Notification{}
	for _, n := range s.notifications {
		if n.Level != level {
			filtered = append(filtered, n)
		}
	}
	s.notifications = filtered
}

// All returns all current notifications.
func (s *NotificationState) All() []Notification {
	return s.notifications
}

// HasAny returns true if there are any notifications.
func (s *NotificationState) HasAny() bool {
	return len(s.notifications) > 0
}

// SetWindowSize updates the window dimensions for positioning calculations.
func (s *NotificationState) SetWindowSize(width, height int) {
	s.windowWidth = width
	s.windowHeight = height
}

// GetLayers creates floating layers for all active notifications.
// Notifications are stacked vertically in the top-right corner of the screen.
func (s *NotificationState) GetLayers(renderFunc func(Notification) string) []*lipgloss.Layer {
	layers := []*lipgloss.Layer{}

	// If window dimensions not set, can't position properly
	if s.windowWidth == 0 {
		return layers
	}

	position := 0
	for i := len(s.notifications) - 1; i >= 0; i-- {
		notification := s.notifications[i]
		notificationView := renderFunc(notification)
		notifWidth := lipgloss.Width(notificationView)
		notifHeight := lipgloss.Height(notificationView)

		// Calculate row based on accumulated heights of previous (newer) notifications
		row := 0
		for j := 0; j < position; j++ {
			idx := len(s.notifications) - 1 - j
			prevNotif := renderFunc(s.notifications[idx])
			row += lipgloss.Height(prevNotif) + 1
		}

		col := max(s.windowWidth-notifWidth-1, 0)
		if row+notifHeight >= s.windowHeight {
			break
		}

		layers = append(layers, lipgloss.NewLayer(notificationView).X(col).Y(row))
		position++
	}

	return layers
}
