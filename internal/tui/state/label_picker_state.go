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
	// Items contains all available labels with their selection states
	Items []LabelPickerItem

	// Cursor is the current cursor position in the picker list
	Cursor int

	// Filter is the text filter for searching labels
	Filter string

	// TaskID is the ID of the task being edited
	TaskID int

	// ColorIdx is the cursor position in the color picker (when creating new labels)
	ColorIdx int

	// CreateMode indicates whether we're in color selection mode for new label creation
	CreateMode bool

	// ReturnMode is the mode to return to after label selection
	ReturnMode Mode
}

// NewLabelPickerState creates a new LabelPickerState with default values.
func NewLabelPickerState() *LabelPickerState {
	return &LabelPickerState{
		Items:      []LabelPickerItem{},
		Cursor:     0,
		Filter:     "",
		TaskID:     0,
		ColorIdx:   0,
		CreateMode: false,
		ReturnMode: NormalMode,
	}
}

// GetFilteredItems returns label picker items filtered by the current filter text.
// If no filter is set, returns all items.
func (s *LabelPickerState) GetFilteredItems() []LabelPickerItem {
	if s.Filter == "" {
		return s.Items
	}

	lowerFilter := strings.ToLower(s.Filter)
	var filtered []LabelPickerItem
	for _, item := range s.Items {
		if strings.Contains(strings.ToLower(item.Label.Name), lowerFilter) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// Clear resets all state to default values.
func (s *LabelPickerState) Clear() {
	s.Items = []LabelPickerItem{}
	s.Cursor = 0
	s.Filter = ""
	s.TaskID = 0
	s.ColorIdx = 0
	s.CreateMode = false
	s.ReturnMode = NormalMode
}

// MoveCursorUp moves the cursor up one position if possible.
// Returns true if the cursor moved, false if already at top.
func (s *LabelPickerState) MoveCursorUp() bool {
	if s.Cursor > 0 {
		s.Cursor--
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
	if s.Cursor < maxIdx {
		s.Cursor++
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

	if len(s.Filter) >= maxFilterLength {
		return false
	}

	s.Filter += string(c)
	return true
}

// BackspaceFilter removes the last character from the filter text.
// Returns true if a character was removed, false if filter was already empty.
func (s *LabelPickerState) BackspaceFilter() bool {
	if len(s.Filter) == 0 {
		return false
	}

	s.Filter = s.Filter[:len(s.Filter)-1]
	return true
}

// UpdateItemSelection updates the selection state of a specific item.
// Returns true if the item was found and updated, false otherwise.
//
// Parameters:
//   - labelID: the ID of the label to update
//   - selected: the new selection state
func (s *LabelPickerState) UpdateItemSelection(labelID int, selected bool) bool {
	for i := range s.Items {
		if s.Items[i].Label.ID == labelID {
			s.Items[i].Selected = selected
			return true
		}
	}
	return false
}

// AddItem adds a new label picker item to the list.
func (s *LabelPickerState) AddItem(item LabelPickerItem) {
	s.Items = append(s.Items, item)
}
