package notifications

// Severity represents the severity level of a notification
type Severity int

const (
	Info Severity = iota
	Warning
	Error
)
