package state

// UIElements groups all UI presentation states into a single struct.
// This reduces the number of fields in the Model and improves organization.
// UI elements manage the presentation and interaction state of various UI components.
type UIElements struct {
	Notification *NotificationState // Notification state (for displaying user messages)
	Search       *SearchState       // Search state (for filtering/searching tasks)
	Filter       *FilterState       // Filter state (for filter bar chip selections)
	ListView     *ListViewState     // List view state (for rendering tasks in list format)
	CurrentTip   string             // Current tip to display in statusbar
}

// NewUIElements creates a new UIElements instance with all UI element states initialized.
func NewUIElements() *UIElements {
	return &UIElements{
		Notification: NewNotificationState(),
		Search:       NewSearchState(),
		Filter:       NewFilterState(),
		ListView:     NewListViewState(),
	}
}
