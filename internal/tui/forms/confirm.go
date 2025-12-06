package forms

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Confirm is a yes/no confirmation field
type Confirm struct {
	key        string
	title      string
	affirmative string
	negative    string
	value      *bool
	focused    bool
	selection  bool // true = yes, false = no
}

// NewConfirm creates a new confirm field
func NewConfirm(key, title, affirmative, negative string, value *bool) *Confirm {
	selection := true
	if value != nil {
		selection = *value
	}

	return &Confirm{
		key:         key,
		title:       title,
		affirmative: affirmative,
		negative:    negative,
		value:       value,
		selection:   selection,
	}
}

// Update handles messages
func (c *Confirm) Update(msg tea.Msg) (Field, tea.Cmd) {
	if !c.focused {
		return c, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "left", "h":
			c.selection = true
		case "right", "l":
			c.selection = false
		case "enter", " ":
			// Toggle selection
			c.selection = !c.selection
		}

		// Update the value pointer
		if c.value != nil {
			*c.value = c.selection
		}
	}

	return c, nil
}

// View renders the confirm field
func (c *Confirm) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")).Background(lipgloss.Color("235"))
	unselectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	title := titleStyle.Render(c.title) + "\n"

	var yesStyle, noStyle lipgloss.Style
	if c.selection {
		yesStyle = selectedStyle
		noStyle = unselectedStyle
	} else {
		yesStyle = unselectedStyle
		noStyle = selectedStyle
	}

	yesOption := yesStyle.Render(" " + c.affirmative + " ")
	noOption := noStyle.Render(" " + c.negative + " ")

	if c.focused {
		return title + yesOption + "  " + noOption
	}
	return title + yesOption + "  " + noOption
}

// Focus focuses the confirm field
func (c *Confirm) Focus() tea.Cmd {
	c.focused = true
	return nil
}

// Blur removes focus
func (c *Confirm) Blur() {
	c.focused = false
}

// Focused returns whether the field is focused
func (c *Confirm) Focused() bool {
	return c.focused
}

// Key returns the field key
func (c *Confirm) Key() string {
	return c.key
}

// Value returns the current selection
func (c *Confirm) Value() bool {
	return c.selection
}
