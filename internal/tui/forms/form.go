package forms

import (
	tea "charm.land/bubbletea/v2"
)

// FormState represents the state of the form
type FormState int

const (
	StateInProgress FormState = iota
	StateCompleted
	StateAborted
)

// Field is the interface that all form fields must implement
type Field interface {
	// Update handles messages and updates the field
	Update(tea.Msg) (Field, tea.Cmd)

	// View renders the field
	View() string

	// Focus focuses the field
	Focus() tea.Cmd

	// Blur removes focus from the field
	Blur()

	// Focused returns whether the field is focused
	Focused() bool

	// Key returns the field's key (used to retrieve values)
	Key() string
}

// Form manages a collection of fields
type Form struct {
	fields       []Field
	focusedIndex int
	state        FormState
}

// NewForm creates a new form with the given fields
func NewForm(fields ...Field) *Form {
	return &Form{
		fields:       fields,
		focusedIndex: 0,
		state:        StateInProgress,
	}
}

// Init initializes the form
func (f *Form) Init() tea.Cmd {
	if len(f.fields) > 0 {
		return f.fields[0].Focus()
	}
	return nil
}

// Update handles messages for the form
func (f *Form) Update(msg tea.Msg) (*Form, tea.Cmd) {
	if f.state != StateInProgress {
		return f, nil
	}

	// Handle keyboard messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			f.state = StateAborted
			return f, nil

		case "tab", "shift+tab":
			return f, f.handleTabNavigation(keyMsg.String() == "shift+tab")

		case "enter":
			// Only submit if on last field and it's not a textarea
			if f.focusedIndex == len(f.fields)-1 {
				// Check if current field allows enter to submit
				// For now, we'll handle this in the specific field types
			}
		}
	}

	// Forward message to focused field
	if f.focusedIndex < len(f.fields) {
		var cmd tea.Cmd
		f.fields[f.focusedIndex], cmd = f.fields[f.focusedIndex].Update(msg)
		return f, cmd
	}

	return f, nil
}

// handleTabNavigation moves focus between fields
func (f *Form) handleTabNavigation(reverse bool) tea.Cmd {
	if len(f.fields) == 0 {
		return nil
	}

	// Blur current field
	f.fields[f.focusedIndex].Blur()

	// Move focus
	if reverse {
		f.focusedIndex--
		if f.focusedIndex < 0 {
			f.focusedIndex = len(f.fields) - 1
		}
	} else {
		f.focusedIndex++
		if f.focusedIndex >= len(f.fields) {
			f.focusedIndex = 0
		}
	}

	// Focus new field
	return f.fields[f.focusedIndex].Focus()
}

// View renders the form
func (f *Form) View() string {
	s := ""
	for _, field := range f.fields {
		s += field.View() + "\n\n"
	}
	return s
}

// State returns the current form state
func (f *Form) State() FormState {
	return f.state
}

// Submit marks the form as completed
func (f *Form) Submit() {
	f.state = StateCompleted
}

// Abort marks the form as aborted
func (f *Form) Abort() {
	f.state = StateAborted
}

// Get retrieves a field by key
func (f *Form) Get(key string) Field {
	for _, field := range f.fields {
		if field.Key() == key {
			return field
		}
	}
	return nil
}
