package state

// EstimateInputState manages the estimate input modal state.
// This modal allows users to enter or clear a time estimate for a task.
type EstimateInputState struct {
	buffer     string
	error      string
	ReturnMode Mode
}

// NewEstimateInputState creates a new EstimateInputState with default values.
func NewEstimateInputState() *EstimateInputState {
	return &EstimateInputState{
		ReturnMode: TaskFormMode,
	}
}

// Buffer returns the current input text.
func (s *EstimateInputState) Buffer() string {
	return s.buffer
}

// SetBuffer sets the input text.
func (s *EstimateInputState) SetBuffer(val string) {
	s.buffer = val
}

// Error returns the current validation error message.
func (s *EstimateInputState) Error() string {
	return s.error
}

// SetError sets the validation error message.
func (s *EstimateInputState) SetError(msg string) {
	s.error = msg
}

// Reset clears all state to default values.
func (s *EstimateInputState) Reset() {
	s.buffer = ""
	s.error = ""
	s.ReturnMode = TaskFormMode
}

// AppendChar adds a character to the buffer and validates the result.
// If validation fails, it sets the error to the provided hint.
// If validation succeeds, it clears any existing error.
func (s *EstimateInputState) AppendChar(ch string, validate func(*string) error, hint string) {
	s.buffer += ch
	if err := validate(&s.buffer); err != nil {
		s.error = hint
	} else {
		s.error = ""
	}
}

// Backspace removes the last character from the buffer and re-validates.
// If the buffer becomes empty, the error is cleared.
// Otherwise, validation is run and the error is set/cleared accordingly.
func (s *EstimateInputState) Backspace(validate func(*string) error, hint string) {
	if len(s.buffer) == 0 {
		return
	}

	s.buffer = s.buffer[:len(s.buffer)-1]

	if s.buffer == "" {
		s.error = ""
	} else if err := validate(&s.buffer); err != nil {
		s.error = hint
	} else {
		s.error = ""
	}
}

// Submit validates the current buffer and returns the result.
// Returns (buffer, true) if the buffer is empty or passes validation.
// Returns ("", false) if validation fails, and sets the error.
func (s *EstimateInputState) Submit(validate func(*string) error, hint string) (string, bool) {
	if s.buffer == "" {
		return "", true
	}

	if err := validate(&s.buffer); err != nil {
		s.error = hint
		return "", false
	}

	return s.buffer, true
}
