package state

// PriorityPickerState manages the priority picker modal state.
// This modal allows users to change the priority of a task.
type PriorityPickerState struct {
	// selectedPriorityID is the currently selected priority ID
	selectedPriorityID int

	// cursor is the current cursor position in the priority picker
	cursor int

	// returnMode is the mode to return to after closing the picker
	returnMode Mode
}

// NewPriorityPickerState creates a new PriorityPickerState with default values.
func NewPriorityPickerState() *PriorityPickerState {
	return &PriorityPickerState{
		selectedPriorityID: 3, // Default to medium (id=3)
		cursor:             0,
		returnMode:         TicketFormMode,
	}
}

// SelectedPriorityID returns the currently selected priority ID.
func (s *PriorityPickerState) SelectedPriorityID() int {
	return s.selectedPriorityID
}

// SetSelectedPriorityID updates the selected priority ID.
func (s *PriorityPickerState) SetSelectedPriorityID(id int) {
	s.selectedPriorityID = id
}

// Cursor returns the current cursor position.
func (s *PriorityPickerState) Cursor() int {
	return s.cursor
}

// SetCursor updates the cursor position.
func (s *PriorityPickerState) SetCursor(idx int) {
	s.cursor = idx
}

// ReturnMode returns the mode to return to after closing.
func (s *PriorityPickerState) ReturnMode() Mode {
	return s.returnMode
}

// SetReturnMode updates the return mode.
func (s *PriorityPickerState) SetReturnMode(mode Mode) {
	s.returnMode = mode
}

// MoveUp moves the cursor up one position if possible.
func (s *PriorityPickerState) MoveUp() {
	if s.cursor > 0 {
		s.cursor--
	}
}

// MoveDown moves the cursor down one position if possible.
// There are 5 priority levels (indices 0-4).
func (s *PriorityPickerState) MoveDown() {
	const maxPriorities = 5
	if s.cursor < maxPriorities-1 {
		s.cursor++
	}
}

// Reset resets all state to default values.
func (s *PriorityPickerState) Reset() {
	s.selectedPriorityID = 3 // Default to medium
	s.cursor = 0
	s.returnMode = TicketFormMode
}
