package state

import "strings"

// InputState manages simple text input state for dialogs.
// This is used for column creation/renaming and other simple text inputs.
// For complex forms with multiple fields, see FormState.
type InputState struct {
	// Buffer contains the text currently being typed
	Buffer string

	// Prompt is the text displayed to the user (e.g., "New column name:")
	Prompt string

	// DeleteColumnTaskCount stores the number of tasks in a column being deleted
	// This is used to show a warning message in the delete confirmation dialog
	DeleteColumnTaskCount int

	// InitialBuffer stores the original buffer value for change detection (EditColumnMode)
	InitialBuffer string
}

// NewInputState creates a new InputState with empty values.
func NewInputState() *InputState {
	return &InputState{
		Buffer:                "",
		Prompt:                "",
		DeleteColumnTaskCount: 0,
	}
}

// Clear resets the buffer and prompt to empty strings.
// The task count is not cleared as it's managed separately.
func (s *InputState) Clear() {
	s.Buffer = ""
	s.Prompt = ""
	s.InitialBuffer = ""
}

// AppendChar appends a character to the input buffer if within max length.
// Returns true if the character was added, false if buffer is at max length.
//
// Parameters:
//   - c: the character to append
//
// The maximum buffer length is currently set to 100 characters to prevent
// excessive memory usage and ensure UI readability.
func (s *InputState) AppendChar(c rune) bool {
	const maxLength = 100

	if len(s.Buffer) >= maxLength {
		return false
	}

	s.Buffer += string(c)
	return true
}

// Backspace removes the last character from the input buffer.
// Returns true if a character was removed, false if buffer was already empty.
func (s *InputState) Backspace() bool {
	if len(s.Buffer) == 0 {
		return false
	}

	s.Buffer = s.Buffer[:len(s.Buffer)-1]
	return true
}

// IsEmpty returns true if the input buffer is empty or contains only whitespace.
func (s *InputState) IsEmpty() bool {
	return len(s.Buffer) == 0 || len(strings.TrimSpace(s.Buffer)) == 0
}

// TrimmedBuffer returns the input buffer with leading and trailing whitespace removed.
func (s *InputState) TrimmedBuffer() string {
	return strings.TrimSpace(s.Buffer)
}

// HasInputChanges returns true if the buffer differs from initial value.
// Used for EditColumnMode to detect changes.
func (s *InputState) HasInputChanges() bool {
	return strings.TrimSpace(s.Buffer) != strings.TrimSpace(s.InitialBuffer)
}

// SnapshotInitialBuffer stores current buffer as initial value.
// Call this when entering EditColumnMode to track the original column name.
func (s *InputState) SnapshotInitialBuffer() {
	s.InitialBuffer = s.Buffer
}
