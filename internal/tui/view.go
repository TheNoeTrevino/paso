package tui

import "github.com/charmbracelet/lipgloss"

// View renders the current state of the application
// This implements the "View" part of the Model-View-Update pattern
func (m Model) View() string {
	// Wait for terminal size to be initialized
	if m.width == 0 {
		return "Loading..."
	}

	// Render each column with its tasks
	var columns []string
	for _, col := range m.columns {
		tasks := m.tasks[col.ID]
		columns = append(columns, RenderColumn(col, tasks))
	}

	// Layout columns horizontally, aligned to top
	board := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	// Create header and footer
	header := TitleStyle.Render("PASO - Your Tasks")
	footer := "Press 'q' to quit"

	// Combine all elements vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		board,
		"",
		footer,
	)
}
