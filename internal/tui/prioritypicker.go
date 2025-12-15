package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// PriorityOption represents a single priority option in the picker
type PriorityOption struct {
	ID          int
	Description string
	Color       string
}

// GetPriorityOptions returns all available priority options
func GetPriorityOptions() []PriorityOption {
	return []PriorityOption{
		{ID: 1, Description: "trivial", Color: "#3B82F6"},
		{ID: 2, Description: "low", Color: "#22C55E"},
		{ID: 3, Description: "medium", Color: "#EAB308"},
		{ID: 4, Description: "high", Color: "#F97316"},
		{ID: 5, Description: "critical", Color: "#EF4444"},
	}
}

// RenderPriorityPicker renders the priority picker popup
func RenderPriorityPicker(
	priorities []PriorityOption,
	selectedPriorityID int,
	cursorIdx int,
	width int,
) string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))
	content.WriteString(titleStyle.Render("Priority") + "\n\n")

	// Priority list
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Normal))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))

	content.WriteString("Select priority level:\n\n")

	// Render each priority option
	for i, priority := range priorities {
		// Selection indicator (radio button style)
		indicator := "( )"
		if priority.ID == selectedPriorityID {
			indicator = "(•)"
		}

		// Color indicator (small colored block)
		colorBlock := lipgloss.NewStyle().
			Background(lipgloss.Color(priority.Color)).
			Render("  ")

		// Priority name with color
		priorityNameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(priority.Color))
		priorityName := priorityNameStyle.Render(priority.Description)

		line := indicator + " " + colorBlock + " " + priorityName

		// Apply cursor styling
		if i == cursorIdx {
			content.WriteString(selectedStyle.Render("> "+line) + "\n")
		} else {
			content.WriteString(normalStyle.Render("  "+line) + "\n")
		}
	}

	// Help text
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("Enter: select  Esc: cancel") + "\n")

	return content.String()
}
