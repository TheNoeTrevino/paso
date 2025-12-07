package state

import (
	"github.com/thenoetrevino/paso/internal/models"
)

// Mode represents the current interaction mode of the TUI.
// Each mode determines which keyboard shortcuts are active and what UI is displayed.
type Mode int

const (
	NormalMode              Mode = iota // Default navigation mode
	DeleteConfirmMode                   // Confirming task deletion
	AddColumnMode                       // Creating a new column
	EditColumnMode                      // Renaming an existing column
	DeleteColumnConfirmMode             // Confirming column deletion
	HelpMode                            // Displaying help screen
	ViewTaskMode                        // Viewing full task details
	TicketFormMode                      // Full ticket form with huh
	ProjectFormMode                     // Creating a new project with huh
	LabelManagementMode                 // Managing labels (create/edit/delete)
	LabelAssignMode                     // Quick label assignment to task
	LabelPickerMode                     // GitHub-style label picker popup
	ParentPickerMode                    // Parent issue picker popup
	ChildPickerMode                     // Child issue picker popup
)

// UIState manages the user interface state.
// This includes navigation (column/task selection), viewport scrolling,
// terminal dimensions, and the current interaction mode.
type UIState struct {
	// selectedColumn is the index of the currently selected column
	selectedColumn int

	// selectedTask is the index of the currently selected task within the selected column
	selectedTask int

	// width is the current terminal width in characters
	width int

	// height is the current terminal height in characters
	height int

	// mode is the current interaction mode
	mode Mode

	// viewportOffset is the index of the leftmost visible column
	viewportOffset int

	// viewportSize is the number of columns that fit on the screen
	viewportSize int

	// viewingTask is the full task detail currently being viewed (nil if not in ViewTaskMode)
	viewingTask *models.TaskDetail
}

// NewUIState creates a new UIState with default values.
func NewUIState() *UIState {
	return &UIState{
		selectedColumn: 0,
		selectedTask:   0,
		width:          0,
		height:         0,
		mode:           NormalMode,
		viewportOffset: 0,
		viewportSize:   1, // Default to 1, will be recalculated when width is set
		viewingTask:    nil,
	}
}

// SelectedColumn returns the index of the currently selected column.
func (s *UIState) SelectedColumn() int {
	return s.selectedColumn
}

// SetSelectedColumn updates the selected column index.
func (s *UIState) SetSelectedColumn(index int) {
	s.selectedColumn = index
}

// SelectedTask returns the index of the currently selected task.
func (s *UIState) SelectedTask() int {
	return s.selectedTask
}

// SetSelectedTask updates the selected task index.
func (s *UIState) SetSelectedTask(index int) {
	s.selectedTask = index
}

// Width returns the current terminal width.
func (s *UIState) Width() int {
	return s.width
}

// SetWidth updates the terminal width and recalculates viewport size.
func (s *UIState) SetWidth(width int) {
	s.width = width
	s.calculateViewportSize()
}

// Height returns the current terminal height.
func (s *UIState) Height() int {
	return s.height
}

// SetHeight updates the terminal height.
func (s *UIState) SetHeight(height int) {
	s.height = height
}

// Mode returns the current interaction mode.
func (s *UIState) Mode() Mode {
	return s.mode
}

// SetMode updates the current interaction mode.
func (s *UIState) SetMode(mode Mode) {
	s.mode = mode
}

// ViewportOffset returns the index of the leftmost visible column.
func (s *UIState) ViewportOffset() int {
	return s.viewportOffset
}

// SetViewportOffset updates the viewport offset.
func (s *UIState) SetViewportOffset(offset int) {
	s.viewportOffset = offset
}

// ViewportSize returns the number of columns that fit on screen.
func (s *UIState) ViewportSize() int {
	return s.viewportSize
}

// ViewingTask returns the task currently being viewed in detail.
func (s *UIState) ViewingTask() *models.TaskDetail {
	return s.viewingTask
}

// SetViewingTask updates the task being viewed.
func (s *UIState) SetViewingTask(task *models.TaskDetail) {
	s.viewingTask = task
}

// calculateViewportSize calculates how many columns can fit in the terminal width.
//
// Column layout:
//   - Content width: 40 characters
//   - Padding: 2 characters (1 on each side)
//   - Border: 2 characters (1 on each side)
//   - Spacing: 2 characters (between columns)
//   - Total per column: 46 characters
//
// The calculation reserves 4 characters for margins and scroll indicators,
// and ensures at least 1 column is always visible.
func (s *UIState) calculateViewportSize() {
	if s.width == 0 {
		s.viewportSize = 1
		return
	}

	const columnWidth = 46 // 40 content + 2 padding + 2 border + 2 spacing
	const reservedWidth = 4 // margins and scroll indicators

	availableWidth := s.width - reservedWidth

	// Calculate how many columns fit, with minimum of 1
	s.viewportSize = max(1, availableWidth/columnWidth)
}

// AdjustViewportAfterColumnRemoval adjusts the viewport offset after a column is removed.
// This ensures the viewport stays within valid bounds and the selection remains visible.
//
// Parameters:
//   - selectedColumn: the current selected column index
//   - columnsLen: the total number of columns after removal
func (s *UIState) AdjustViewportAfterColumnRemoval(selectedColumn, columnsLen int) {
	if columnsLen == 0 {
		s.viewportOffset = 0
		return
	}

	// If selected column is before viewport, move viewport left
	if selectedColumn < s.viewportOffset {
		s.viewportOffset = selectedColumn
	}

	// If viewport offset is now beyond available columns, adjust it
	if s.viewportOffset+s.viewportSize > columnsLen {
		s.viewportOffset = max(0, columnsLen-s.viewportSize)
	}
}

// ScrollViewportLeft scrolls the viewport one column to the left.
// Returns true if scrolling occurred, false if already at leftmost position.
func (s *UIState) ScrollViewportLeft() bool {
	if s.viewportOffset > 0 {
		s.viewportOffset--
		return true
	}
	return false
}

// ScrollViewportRight scrolls the viewport one column to the right.
// Returns true if scrolling occurred, false if already at rightmost position.
//
// Parameters:
//   - columnsLen: the total number of columns
func (s *UIState) ScrollViewportRight(columnsLen int) bool {
	if s.viewportOffset+s.viewportSize < columnsLen {
		s.viewportOffset++
		return true
	}
	return false
}

// EnsureSelectionVisible adjusts the viewport to ensure the selected column is visible.
// This should be called after navigation or when the selection changes.
func (s *UIState) EnsureSelectionVisible(selectedColumn int) {
	// If selection is off-screen to the left, scroll left
	if selectedColumn < s.viewportOffset {
		s.viewportOffset = selectedColumn
	}

	// If selection is off-screen to the right, scroll right
	if selectedColumn >= s.viewportOffset+s.viewportSize {
		s.viewportOffset = selectedColumn - s.viewportSize + 1
	}
}

// ResetSelection resets both column and task selection to zero.
// This is typically called when switching projects or clearing state.
func (s *UIState) ResetSelection() {
	s.selectedColumn = 0
	s.selectedTask = 0
	s.viewportOffset = 0
}

