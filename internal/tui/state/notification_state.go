package state

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
}

// NewNotificationState creates a new NotificationState with no notifications.
func NewNotificationState() *NotificationState {
	return &NotificationState{
		notifications: []Notification{},
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
