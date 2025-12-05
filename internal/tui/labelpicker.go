package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// RenderLabelPicker renders the label picker popup
// This is a GitHub-style label picker with checkboxes
// Note: filteredItems should already be filtered by the caller using LabelPickerState.GetFilteredItems()
func RenderLabelPicker(
	filteredItems []state.LabelPickerItem,
	cursorIdx int,
	filterText string,
	showCreateOption bool,
	width int,
	height int,
) string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170"))
	content.WriteString(titleStyle.Render("Labels") + "\n\n")

	// Filter input
	filterStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(width - 8)

	filterDisplay := filterText
	if filterDisplay == "" {
		filterDisplay = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Filter labels...")
	}
	content.WriteString(filterStyle.Render(filterDisplay) + "\n\n")

	// Label list
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// Render each label item (already filtered by caller)
	for i, item := range filteredItems {
		// Checkbox indicator
		checkbox := "[ ]"
		if item.Selected {
			checkbox = "[x]"
		}

		// Color indicator (small colored block)
		colorBlock := lipgloss.NewStyle().
			Background(lipgloss.Color(item.Label.Color)).
			Render("  ")

		// Label name
		line := checkbox + " " + colorBlock + " " + item.Label.Name

		// Apply cursor styling
		if i == cursorIdx {
			content.WriteString(selectedStyle.Render("> " + line) + "\n")
		} else {
			content.WriteString(normalStyle.Render("  " + line) + "\n")
		}
	}

	// Show "no results" if filter matched nothing
	if len(filteredItems) == 0 && filterText != "" {
		content.WriteString(dimStyle.Render("  No labels match \"" + filterText + "\"") + "\n")
	}

	// Create new label option
	if showCreateOption {
		content.WriteString("\n")
		createOptionIdx := len(filteredItems)
		createText := "+ Create new label"
		if filterText != "" {
			createText = "+ Create \"" + filterText + "\""
		}

		if cursorIdx == createOptionIdx {
			content.WriteString(selectedStyle.Render("> " + createText) + "\n")
		} else {
			content.WriteString(normalStyle.Render("  " + createText) + "\n")
		}
	}

	// Help text
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("Enter: toggle/create  Esc: close") + "\n")

	return content.String()
}

// RenderLabelColorPicker renders a color picker for creating/editing labels
func RenderLabelColorPicker(
	colors []struct {
		Name  string
		Color string
	},
	cursorIdx int,
	labelName string,
	width int,
) string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")) // Green for creation
	content.WriteString(titleStyle.Render("New Label: "+labelName) + "\n\n")

	// Color options
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	content.WriteString("Select a color:\n\n")

	for i, c := range colors {
		// Color block
		colorBlock := lipgloss.NewStyle().
			Background(lipgloss.Color(c.Color)).
			Render("    ")

		line := colorBlock + " " + c.Name

		if i == cursorIdx {
			content.WriteString(selectedStyle.Render("> " + line) + "\n")
		} else {
			content.WriteString(normalStyle.Render("  " + line) + "\n")
		}
	}

	// Help text
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("Enter: select  Esc: cancel") + "\n")

	return content.String()
}

// GetDefaultLabelColors returns the predefined color options for labels
func GetDefaultLabelColors() []struct {
	Name  string
	Color string
} {
	return []struct {
		Name  string
		Color string
	}{
		{"Red", "#EF4444"},
		{"Orange", "#F97316"},
		{"Yellow", "#EAB308"},
		{"Green", "#22C55E"},
		{"Cyan", "#06B6D4"},
		{"Blue", "#3B82F6"},
		{"Purple", "#7D56F4"},
		{"Pink", "#EC4899"},
		{"Gray", "#6B7280"},
	}
}
