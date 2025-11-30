package state

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
	return len(s.Buffer) == 0 || len(trimSpace(s.Buffer)) == 0
}

// TrimmedBuffer returns the input buffer with leading and trailing whitespace removed.
func (s *InputState) TrimmedBuffer() string {
	return trimSpace(s.Buffer)
}

// trimSpace removes leading and trailing whitespace from a string.
// This is a simple implementation to avoid importing strings package.
func trimSpace(s string) string {
	// Find first non-space character
	start := 0
	for start < len(s) && isSpace(s[start]) {
		start++
	}

	// Find last non-space character
	end := len(s)
	for end > start && isSpace(s[end-1]) {
		end--
	}

	return s[start:end]
}

// isSpace checks if a byte is a whitespace character.
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
