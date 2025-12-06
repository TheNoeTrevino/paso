package forms

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// TextInput is a single-line text input field
type TextInput struct {
	key         string
	title       string
	placeholder string
	value       *string
	input       textinput.Model
}

// NewTextInput creates a new text input field
func NewTextInput(key, title, placeholder string, value *string) *TextInput {
	ti := textinput.New()
	ti.Placeholder = placeholder
	if value != nil && *value != "" {
		ti.SetValue(*value)
	}

	return &TextInput{
		key:         key,
		title:       title,
		placeholder: placeholder,
		value:       value,
		input:       ti,
	}
}

// Update handles messages
func (t *TextInput) Update(msg tea.Msg) (Field, tea.Cmd) {
	var cmd tea.Cmd
	t.input, cmd = t.input.Update(msg)

	// Update the value pointer
	if t.value != nil {
		*t.value = t.input.Value()
	}

	return t, cmd
}

// View renders the text input
func (t *TextInput) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	return titleStyle.Render(t.title) + "\n" + t.input.View()
}

// Focus focuses the text input
func (t *TextInput) Focus() tea.Cmd {
	return t.input.Focus()
}

// Blur removes focus
func (t *TextInput) Blur() {
	t.input.Blur()
}

// Focused returns whether the input is focused
func (t *TextInput) Focused() bool {
	return t.input.Focused()
}

// Key returns the field key
func (t *TextInput) Key() string {
	return t.key
}

// Value returns the current value
func (t *TextInput) Value() string {
	return t.input.Value()
}
