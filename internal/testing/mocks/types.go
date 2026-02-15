package mocks

// MockCall records a single method call to a mock. This type is shared across
// all mock implementations in this package.
type MockCall struct {
	Method string
	TaskID int // Optional: 0 means not applicable. Used by task-related mocks for ergonomic access.
	Args   map[string]any
}
