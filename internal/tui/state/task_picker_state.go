// Package state contains state management structures for the TUI.
// This package provides state objects that track application data and UI state
// without implementing business logic or rendering.
package state

import (
	"fmt"
	"strings"

	"github.com/thenoetrevino/paso/internal/models"
)

// TaskPickerItem represents a single item in the task picker list.
// It combines a task reference with its selection state.
type TaskPickerItem struct {
	// TaskRef is the task reference data from the database
	TaskRef *models.TaskReference

	// Selected indicates whether this task is currently selected
	Selected bool
}

// TaskPickerState manages the task picker popup state.
// This includes the list of available tasks, cursor position, filtering,
// and the picker type (parent/child).
type TaskPickerState struct {
	// Items contains all available tasks with their selection states
	Items []TaskPickerItem

	// Cursor is the current cursor position in the picker list
	Cursor int

	// Filter is the text filter for searching tasks
	Filter string

	// TaskID is the ID of the task being edited
	TaskID int

	// PickerType indicates whether this is a parent or child picker
	PickerType string

	// ReturnMode specifies which mode to return to when closing picker
	// Can be ViewTaskMode or TicketFormMode
	ReturnMode Mode
}

// NewTaskPickerState creates a new TaskPickerState with default values.
func NewTaskPickerState() *TaskPickerState {
	return &TaskPickerState{
		Items:      []TaskPickerItem{},
		Cursor:     0,
		Filter:     "",
		TaskID:     0,
		PickerType: "",
		ReturnMode: NormalMode,
	}
}

// GetFilteredItems returns task picker items filtered by the current filter text.
// If no filter is set, returns all items.
// Filters by both ticket number (PROJ-123 format) and title (case-insensitive substring match).
func (s *TaskPickerState) GetFilteredItems() []TaskPickerItem {
	if s.Filter == "" {
		return s.Items
	}

	lowerFilter := strings.ToLower(s.Filter)
	var filtered []TaskPickerItem
	for _, item := range s.Items {
		// Match on ticket number (PROJ-123 format)
		ticketNum := fmt.Sprintf("%s-%d", item.TaskRef.ProjectName, item.TaskRef.TicketNumber)
		if strings.Contains(strings.ToLower(ticketNum), lowerFilter) {
			filtered = append(filtered, item)
			continue
		}

		// Match on title
		if strings.Contains(strings.ToLower(item.TaskRef.Title), lowerFilter) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// Clear resets all state to default values.
func (s *TaskPickerState) Clear() {
	s.Items = []TaskPickerItem{}
	s.Cursor = 0
	s.Filter = ""
	s.TaskID = 0
	s.PickerType = ""
	s.ReturnMode = NormalMode
}

// MoveCursorUp moves the cursor up one position if possible.
// Returns true if the cursor moved, false if already at top.
func (s *TaskPickerState) MoveCursorUp() bool {
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
func (s *TaskPickerState) MoveCursorDown(maxIdx int) bool {
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
func (s *TaskPickerState) AppendFilter(c rune) bool {
	const maxFilterLength = 50

	if len(s.Filter) >= maxFilterLength {
		return false
	}

	s.Filter += string(c)
	return true
}

// BackspaceFilter removes the last character from the filter text.
// Returns true if a character was removed, false if filter was already empty.
func (s *TaskPickerState) BackspaceFilter() bool {
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
//   - taskID: the ID of the task to update
//   - selected: the new selection state
func (s *TaskPickerState) UpdateItemSelection(taskID int, selected bool) bool {
	for i := range s.Items {
		if s.Items[i].TaskRef.ID == taskID {
			s.Items[i].Selected = selected
			return true
		}
	}
	return false
}

// AddItem adds a new task picker item to the list.
func (s *TaskPickerState) AddItem(item TaskPickerItem) {
	s.Items = append(s.Items, item)
}
