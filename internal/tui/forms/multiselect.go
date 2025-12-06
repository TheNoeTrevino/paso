package forms

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Option represents a selectable option
type Option struct {
	Label string
	Value int
}

// MultiSelect is a multi-select field
type MultiSelect struct {
	key      string
	title    string
	options  []Option
	value    *[]int
	focused  bool
	cursor   int
	selected map[int]bool // tracks which values are selected
}

// NewMultiSelect creates a new multi-select field
func NewMultiSelect(key, title string, options []Option, value *[]int) *MultiSelect {
	selected := make(map[int]bool)
	if value != nil {
		for _, v := range *value {
			selected[v] = true
		}
	}

	return &MultiSelect{
		key:      key,
		title:    title,
		options:  options,
		value:    value,
		selected: selected,
		cursor:   0,
	}
}

// Update handles messages
func (m *MultiSelect) Update(msg tea.Msg) (Field, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case " ", "enter":
			// Toggle selection
			if m.cursor < len(m.options) {
				val := m.options[m.cursor].Value
				if m.selected[val] {
					delete(m.selected, val)
				} else {
					m.selected[val] = true
				}

				// Update the value pointer
				if m.value != nil {
					*m.value = m.getSelectedValues()
				}
			}
		}
	}

	return m, nil
}

// getSelectedValues returns a slice of selected values
func (m *MultiSelect) getSelectedValues() []int {
	var values []int
	for val := range m.selected {
		values = append(values, val)
	}
	return values
}

// View renders the multi-select field
func (m *MultiSelect) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	s := titleStyle.Render(m.title) + "\n"

	for i, option := range m.options {
		cursor := "  "
		if m.focused && i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}

		checkbox := "[ ] "
		if m.selected[option.Value] {
			checkbox = "[x] "
		}

		var labelStyle lipgloss.Style
		if m.selected[option.Value] {
			labelStyle = selectedStyle
		} else {
			labelStyle = normalStyle
		}

		s += cursor + checkbox + labelStyle.Render(option.Label) + "\n"
	}

	return s
}

// Focus focuses the multi-select field
func (m *MultiSelect) Focus() tea.Cmd {
	m.focused = true
	return nil
}

// Blur removes focus
func (m *MultiSelect) Blur() {
	m.focused = false
}

// Focused returns whether the field is focused
func (m *MultiSelect) Focused() bool {
	return m.focused
}

// Key returns the field key
func (m *MultiSelect) Key() string {
	return m.key
}

// SelectedValues returns the currently selected values
func (m *MultiSelect) SelectedValues() []int {
	return m.getSelectedValues()
}
