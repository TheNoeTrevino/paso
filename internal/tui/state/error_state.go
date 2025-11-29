package state

// ErrorState manages error display state.
// This provides a centralized way to handle user-facing error messages
// throughout the application.
type ErrorState struct {
	// message contains the current error message to display
	// Empty string indicates no error
	message string
}

// NewErrorState creates a new ErrorState with no error.
func NewErrorState() *ErrorState {
	return &ErrorState{
		message: "",
	}
}

// Set sets the error message to be displayed.
//
// Parameters:
//   - msg: the error message to display
func (s *ErrorState) Set(msg string) {
	s.message = msg
}

// Clear removes any current error message.
func (s *ErrorState) Clear() {
	s.message = ""
}

// HasError returns true if there is currently an error message set.
func (s *ErrorState) HasError() bool {
	return s.message != ""
}

// Get returns the current error message.
// Returns an empty string if there is no error.
func (s *ErrorState) Get() string {
	return s.message
}
