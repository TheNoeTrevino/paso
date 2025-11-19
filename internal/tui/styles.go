package tui

import "github.com/charmbracelet/lipgloss"

// Style definitions for the kanban board UI
// These styles follow Lipgloss conventions for composable terminal styling

var (
	// ColumnStyle defines the appearance of kanban board columns
	ColumnStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")). // Blue border
		Padding(1).
		Width(30)

	// TaskStyle defines the appearance of individual tasks
	TaskStyle = lipgloss.NewStyle().
		Padding(0, 1).
		MarginBottom(1)

	// TitleStyle defines the appearance of titles (column names, app header)
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")) // Purple text
)
