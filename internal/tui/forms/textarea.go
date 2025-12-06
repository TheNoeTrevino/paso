package forms

import (
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// TextArea is a multi-line text input field
type TextArea struct {
	key         string
	title       string
	placeholder string
	charLimit   int
	value       *string
	textarea    textarea.Model
}

// NewTextArea creates a new text area field
func NewTextArea(key, title, placeholder string, charLimit int, value *string) *TextArea {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.CharLimit = charLimit
	ta.SetHeight(5)
	if value != nil && *value != "" {
		ta.SetValue(*value)
	}

	return &TextArea{
		key:         key,
		title:       title,
		placeholder: placeholder,
		charLimit:   charLimit,
		value:       value,
		textarea:    ta,
	}
}

// Update handles messages
func (t *TextArea) Update(msg tea.Msg) (Field, tea.Cmd) {
	var cmd tea.Cmd
	t.textarea, cmd = t.textarea.Update(msg)

	// Update the value pointer
	if t.value != nil {
		*t.value = t.textarea.Value()
	}

	return t, cmd
}

// View renders the text area
func (t *TextArea) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	return titleStyle.Render(t.title) + "\n" + t.textarea.View()
}

// Focus focuses the text area
func (t *TextArea) Focus() tea.Cmd {
	return t.textarea.Focus()
}

// Blur removes focus
func (t *TextArea) Blur() {
	t.textarea.Blur()
}

// Focused returns whether the textarea is focused
func (t *TextArea) Focused() bool {
	return t.textarea.Focused()
}

// Key returns the field key
func (t *TextArea) Key() string {
	return t.key
}

// Value returns the current value
func (t *TextArea) Value() string {
	return t.textarea.Value()
}
