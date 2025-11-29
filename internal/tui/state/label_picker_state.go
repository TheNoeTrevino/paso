package state

import (
	"strings"

	"github.com/thenoetrevino/paso/internal/models"
)

// LabelPickerItem represents a single item in the label picker list.
// It combines a label with its selection state for the current task.
type LabelPickerItem struct {
	// Label is the label data from the database
	Label *models.Label

	// Selected indicates whether this label is currently assigned to the task
	Selected bool
}

// LabelPickerState manages the GitHub-style label picker popup state.
// This includes the list of available labels, cursor position, filtering,
// and color picker for creating new labels.
type LabelPickerState struct {
	// items contains all available labels with their selection states
	items []LabelPickerItem

	// cursor is the current cursor position in the picker list
	cursor int

	// filter is the text filter for searching labels
	filter string

	// taskID is the ID of the task being edited
	taskID int

	// colorIdx is the cursor position in the color picker (when creating new labels)
	colorIdx int

	// createMode indicates whether we're in color selection mode for new label creation
	createMode bool
}

// NewLabelPickerState creates a new LabelPickerState with default values.
func NewLabelPickerState() *LabelPickerState {
	return &LabelPickerState{
		items:      []LabelPickerItem{},
		cursor:     0,
		filter:     "",
		taskID:     0,
		colorIdx:   0,
		createMode: false,
	}
}

// Items returns the label picker items.
func (s *LabelPickerState) Items() []LabelPickerItem {
	return s.items
}

// SetItems sets the label picker items.
func (s *LabelPickerState) SetItems(items []LabelPickerItem) {
	s.items = items
}

// Cursor returns the current cursor position.
func (s *LabelPickerState) Cursor() int {
	return s.cursor
}

// SetCursor sets the cursor position.
func (s *LabelPickerState) SetCursor(pos int) {
	s.cursor = pos
}

// Filter returns the current filter text.
func (s *LabelPickerState) Filter() string {
	return s.filter
}

// SetFilter sets the filter text.
func (s *LabelPickerState) SetFilter(filter string) {
	s.filter = filter
}

// TaskID returns the ID of the task being edited.
func (s *LabelPickerState) TaskID() int {
	return s.taskID
}

// SetTaskID sets the task ID.
func (s *LabelPickerState) SetTaskID(id int) {
	s.taskID = id
}

// ColorIdx returns the color picker cursor position.
func (s *LabelPickerState) ColorIdx() int {
	return s.colorIdx
}

// SetColorIdx sets the color picker cursor position.
func (s *LabelPickerState) SetColorIdx(idx int) {
	s.colorIdx = idx
}

// CreateMode returns whether we're in create mode.
func (s *LabelPickerState) CreateMode() bool {
	return s.createMode
}

// SetCreateMode sets the create mode state.
func (s *LabelPickerState) SetCreateMode(enabled bool) {
	s.createMode = enabled
}

// GetFilteredItems returns label picker items filtered by the current filter text.
// If no filter is set, returns all items.
func (s *LabelPickerState) GetFilteredItems() []LabelPickerItem {
	if s.filter == "" {
		return s.items
	}

	lowerFilter := strings.ToLower(s.filter)
	var filtered []LabelPickerItem
	for _, item := range s.items {
		if strings.Contains(strings.ToLower(item.Label.Name), lowerFilter) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// Clear resets all state to default values.
func (s *LabelPickerState) Clear() {
	s.items = []LabelPickerItem{}
	s.cursor = 0
	s.filter = ""
	s.taskID = 0
	s.colorIdx = 0
	s.createMode = false
}

// MoveCursorUp moves the cursor up one position if possible.
// Returns true if the cursor moved, false if already at top.
func (s *LabelPickerState) MoveCursorUp() bool {
	if s.cursor > 0 {
		s.cursor--
		return true
	}
	return false
}

// MoveCursorDown moves the cursor down one position if possible.
// Returns true if the cursor moved, false if already at bottom.
//
// Parameters:
//   - maxIdx: the maximum valid cursor position (typically len(items))
func (s *LabelPickerState) MoveCursorDown(maxIdx int) bool {
	if s.cursor < maxIdx {
		s.cursor++
		return true
	}
	return false
}

// AppendFilter appends a character to the filter text.
// Returns true if the character was added, false if filter is at max length.
//
// Parameters:
//   - c: the character to append
func (s *LabelPickerState) AppendFilter(c rune) bool {
	const maxFilterLength = 50

	if len(s.filter) >= maxFilterLength {
		return false
	}

	s.filter += string(c)
	return true
}

// BackspaceFilter removes the last character from the filter text.
// Returns true if a character was removed, false if filter was already empty.
func (s *LabelPickerState) BackspaceFilter() bool {
	if len(s.filter) == 0 {
		return false
	}

	s.filter = s.filter[:len(s.filter)-1]
	return true
}

// UpdateItemSelection updates the selection state of a specific item.
// Returns true if the item was found and updated, false otherwise.
//
// Parameters:
//   - labelID: the ID of the label to update
//   - selected: the new selection state
func (s *LabelPickerState) UpdateItemSelection(labelID int, selected bool) bool {
	for i := range s.items {
		if s.items[i].Label.ID == labelID {
			s.items[i].Selected = selected
			return true
		}
	}
	return false
}

// AddItem adds a new label picker item to the list.
func (s *LabelPickerState) AddItem(item LabelPickerItem) {
	s.items = append(s.items, item)
}
