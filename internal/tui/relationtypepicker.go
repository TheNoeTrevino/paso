package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// RelationTypeOption represents a single relation type option in the picker
type RelationTypeOption struct {
	ID         int
	PToCLabel  string // Label from parent's perspective
	CToPLabel  string // Label from child's perspective
	Color      string
	IsBlocking bool
}

// GetRelationTypeOptions returns all available relation type options
// These are system-defined and hardcoded (no CRUD)
func GetRelationTypeOptions() []RelationTypeOption {
	return []RelationTypeOption{
		{ID: 1, PToCLabel: "Parent", CToPLabel: "Child", Color: "#6B7280", IsBlocking: false},
		{ID: 2, PToCLabel: "Blocked By", CToPLabel: "Blocker", Color: "#EF4444", IsBlocking: true},
		{ID: 3, PToCLabel: "Related To", CToPLabel: "Related To", Color: "#3B82F6", IsBlocking: false},
	}
}

// RenderRelationTypePicker renders the relation type picker popup
func RenderRelationTypePicker(
	relationTypes []RelationTypeOption,
	selectedRelationTypeID int,
	cursorIdx int,
	width int,
	pickerType string, // "parent" or "child" to determine which label to show
) string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))
	content.WriteString(titleStyle.Render("Relation Type") + "\n\n")

	// Relation type list
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Normal))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))

	content.WriteString("Select relation type:\n\n")

	// Render each relation type option
	for i, relationType := range relationTypes {
		// Selection indicator (radio button style)
		indicator := "( )"
		if relationType.ID == selectedRelationTypeID {
			indicator = "(•)"
		}

		// Color indicator (small colored block)
		colorBlock := lipgloss.NewStyle().
			Background(lipgloss.Color(relationType.Color)).
			Render("  ")

		// Determine which label to display based on picker type
		var label string
		if pickerType == "parent" {
			// For parent picker, show how the child sees this relationship
			label = relationType.PToCLabel
		} else {
			// For child picker, show how the parent sees this relationship
			label = relationType.CToPLabel
		}

		// Relation type name with color
		relationNameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(relationType.Color))
		relationName := relationNameStyle.Render(label)

		line := indicator + " " + colorBlock + " " + relationName

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
