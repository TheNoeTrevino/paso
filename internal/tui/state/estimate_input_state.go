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
		ReturnMode: TicketFormMode,
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
	s.ReturnMode = TicketFormMode
}
