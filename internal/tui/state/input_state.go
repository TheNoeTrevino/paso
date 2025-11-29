package state

// InputState manages simple text input state for dialogs.
// This is used for column creation/renaming and other simple text inputs.
// For complex forms with multiple fields, see FormState.
type InputState struct {
	// buffer contains the text currently being typed
	buffer string

	// prompt is the text displayed to the user (e.g., "New column name:")
	prompt string

	// deleteColumnTaskCount stores the number of tasks in a column being deleted
	// This is used to show a warning message in the delete confirmation dialog
	deleteColumnTaskCount int
}

// NewInputState creates a new InputState with empty values.
func NewInputState() *InputState {
	return &InputState{
		buffer:                "",
		prompt:                "",
		deleteColumnTaskCount: 0,
	}
}

// Buffer returns the current input buffer text.
func (s *InputState) Buffer() string {
	return s.buffer
}

// SetBuffer sets the input buffer to the specified text.
func (s *InputState) SetBuffer(text string) {
	s.buffer = text
}

// Prompt returns the current prompt text.
func (s *InputState) Prompt() string {
	return s.prompt
}

// SetPrompt sets the prompt text.
func (s *InputState) SetPrompt(prompt string) {
	s.prompt = prompt
}

// DeleteColumnTaskCount returns the task count for column deletion warning.
func (s *InputState) DeleteColumnTaskCount() int {
	return s.deleteColumnTaskCount
}

// SetDeleteColumnTaskCount sets the task count for deletion warning.
func (s *InputState) SetDeleteColumnTaskCount(count int) {
	s.deleteColumnTaskCount = count
}

// Clear resets the buffer and prompt to empty strings.
// The task count is not cleared as it's managed separately.
func (s *InputState) Clear() {
	s.buffer = ""
	s.prompt = ""
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

	if len(s.buffer) >= maxLength {
		return false
	}

	s.buffer += string(c)
	return true
}

// Backspace removes the last character from the input buffer.
// Returns true if a character was removed, false if buffer was already empty.
func (s *InputState) Backspace() bool {
	if len(s.buffer) == 0 {
		return false
	}

	s.buffer = s.buffer[:len(s.buffer)-1]
	return true
}

// IsEmpty returns true if the input buffer is empty or contains only whitespace.
func (s *InputState) IsEmpty() bool {
	return len(s.buffer) == 0 || len(trimSpace(s.buffer)) == 0
}

// TrimmedBuffer returns the input buffer with leading and trailing whitespace removed.
func (s *InputState) TrimmedBuffer() string {
	return trimSpace(s.buffer)
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
