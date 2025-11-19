package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all messages and updates the model accordingly
// This implements the "Update" part of the Model-View-Update pattern
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// Handle keyboard input
		switch msg.String() {
		case "q", "ctrl+c":
			// Quit the application
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		// Handle terminal resize events
		m.width = msg.Width
		m.height = msg.Height
	}

	// Return updated model with no command
	return m, nil
}
