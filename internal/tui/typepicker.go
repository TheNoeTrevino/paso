package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// TypeOption represents a single type option in the picker
type TypeOption struct {
	ID          int
	Description string
	Color       string
}

// GetTypeOptions returns all available type options
func GetTypeOptions() []TypeOption {
	return []TypeOption{
		{ID: 1, Description: "task", Color: "#3B82F6"},    // Blue
		{ID: 2, Description: "feature", Color: "#A855F7"}, // Purple
	}
}

// RenderTypePicker renders the type picker popup
func RenderTypePicker(
	types []TypeOption,
	selectedTypeID int,
	cursorIdx int,
	width int,
) string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))
	content.WriteString(titleStyle.Render("Type") + "\n\n")

	// Type list
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Normal))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))

	content.WriteString("Select issue type:\n\n")

	// Render each type option
	for i, typeOpt := range types {
		// Selection indicator (radio button style)
		indicator := "( )"
		if typeOpt.ID == selectedTypeID {
			indicator = "(•)"
		}

		// Color indicator (small colored block)
		colorBlock := lipgloss.NewStyle().
			Background(lipgloss.Color(typeOpt.Color)).
			Render("  ")

		// Type name with color
		typeNameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(typeOpt.Color))
		typeName := typeNameStyle.Render(typeOpt.Description)

		line := indicator + " " + colorBlock + " " + typeName

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
