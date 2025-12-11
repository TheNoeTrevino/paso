package state

import "charm.land/lipgloss/v2"

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

// Notification represents a single notification message with a severity level.
type Notification struct {
	Level   NotificationLevel
	Message string
}

// NotificationState manages notification display state.
// This provides a centralized way to handle user-facing notifications
// of different severity levels throughout the application.
type NotificationState struct {
	// notifications contains the list of current notifications to display
	notifications []Notification
	// windowWidth tracks the current window width for positioning
	windowWidth int
	// windowHeight tracks the current window height for positioning
	windowHeight int
}

// NewNotificationState creates a new NotificationState with no notifications.
func NewNotificationState() *NotificationState {
	return &NotificationState{
		notifications: []Notification{},
		windowWidth:   0,
		windowHeight:  0,
	}
}

// Add adds a new notification with the specified level and message.
//
// Parameters:
//   - level: the severity level of the notification
//   - message: the notification message to display
func (s *NotificationState) Add(level NotificationLevel, message string) {
	s.notifications = append(s.notifications, Notification{
		Level:   level,
		Message: message,
	})
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

	for i, notification := range s.notifications {
		notificationView := renderFunc(notification)
		notifWidth := lipgloss.Width(notificationView)
		notifHeight := lipgloss.Height(notificationView)

		// Stack vertically from top-right
		// Calculate row based on accumulated heights of previous notifications
		row := 0
		for j := 0; j < i; j++ {
			prevNotif := renderFunc(s.notifications[j])
			row += lipgloss.Height(prevNotif) + 1 // +1 for spacing
		}

		col := s.windowWidth - notifWidth - 1 // 1 char padding from right edge

		// Ensure we don't go off screen
		if col < 0 {
			col = 0
		}
		if row+notifHeight >= s.windowHeight {
			// Don't render notifications that would go off screen
			break
		}

		layers = append(layers,
			lipgloss.NewLayer(notificationView).X(col).Y(row))
	}

	return layers
}
