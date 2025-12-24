package renderers

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
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
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))
	content.WriteString(titleStyle.Render("Labels") + "\n\n")

	// Filter input
	filterStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.Subtle)).
		Padding(0, 1).
		Width(width - 8)

	filterDisplay := filterText
	if filterDisplay == "" {
		filterDisplay = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Render("Filter labels...")
	}
	content.WriteString(filterStyle.Render(filterDisplay) + "\n\n")

	// Label list
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Normal))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))

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
			content.WriteString(selectedStyle.Render("> "+line) + "\n")
		} else {
			content.WriteString(normalStyle.Render("  "+line) + "\n")
		}
	}

	// Show "no results" if filter matched nothing
	if len(filteredItems) == 0 && filterText != "" {
		content.WriteString(dimStyle.Render("  No labels match \""+filterText+"\"") + "\n")
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
			content.WriteString(selectedStyle.Render("> "+createText) + "\n")
		} else {
			content.WriteString(normalStyle.Render("  "+createText) + "\n")
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
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Create))
	content.WriteString(titleStyle.Render("New Label: "+labelName) + "\n\n")

	// Color options
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Normal))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))

	content.WriteString("Select a color:\n\n")

	for i, c := range colors {
		// Color block
		colorBlock := lipgloss.NewStyle().
			Background(lipgloss.Color(c.Color)).
			Render("    ")

		line := colorBlock + " " + c.Name

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

// RenderTaskPicker renders the task picker popup in a GitHub-style interface with checkboxes.
// This picker is used for both parent and child issue selection.
//
// The picker displays tasks in PROJ-123 format with titles, allows filtering by typing,
// and indicates selected tasks with checkboxes. The cursor position determines which
// task is currently highlighted.
//
// Parameters:
//   - filteredItems: Pre-filtered list of tasks to display. The caller must filter using
//     TaskPickerState.GetFilteredItems() before calling this function.
//   - cursorIdx: Zero-based index of the currently highlighted item in filteredItems.
//   - filterText: Current filter text entered by the user. If empty, shows placeholder text.
//   - pickerType: Title displayed at the top ("Parent Issues" or "Child Issues").
//   - width: Maximum width for the picker content in characters.
//   - height: Maximum height for the picker content in rows (currently unused but reserved for scrolling).
//   - isParentPicker: True for parent picker, false for child picker. Determines which label to show.
//
// Returns:
//   - A formatted string ready for rendering in the terminal.
func RenderTaskPicker(
	filteredItems []state.TaskPickerItem,
	cursorIdx int,
	filterText string,
	pickerType string, // "Parent Issues" or "Child Issues"
	width int,
	height int,
	isParentPicker bool,
) string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))
	content.WriteString(titleStyle.Render(pickerType) + "\n\n")

	// Filter input
	filterStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.Subtle)).
		Padding(0, 1).
		Width(width - 8)

	filterDisplay := filterText
	if filterDisplay == "" {
		filterDisplay = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Render("Filter tasks...")
	}
	content.WriteString(filterStyle.Render(filterDisplay) + "\n\n")

	// Task list
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Normal))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))

	// Render each task item (already filtered by caller)
	for i, item := range filteredItems {
		// Checkbox indicator
		checkbox := "[ ]"
		if item.Selected {
			checkbox = "[x]"
		}

		// Task identifier (PROJ-123 format)
		taskID := fmt.Sprintf("%s-%d", item.TaskRef.ProjectName, item.TaskRef.TicketNumber)

		// For selected items, show relation type
		var relationLabel string
		var relationLabelPlain string // For length calculation
		if item.Selected && item.RelationTypeID > 0 {
			relationTypes := GetRelationTypeOptions()
			for _, rt := range relationTypes {
				if rt.ID == item.RelationTypeID {
					label := rt.CToPLabel
					if isParentPicker {
						label = rt.PToCLabel
					}
					labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(rt.Color))
					relationLabel = labelStyle.Render(label)
					relationLabelPlain = label
					break
				}
			}
		}

		// Truncate title if needed to fit width (reserve space for relation label)
		maxLen := width - 12
		if relationLabelPlain != "" {
			maxLen -= len(relationLabelPlain) + 4 // Reserve space for label + padding
		}
		title := item.TaskRef.Title
		if len(title) > maxLen {
			title = title[:maxLen-3] + "..."
		}

		// Build line: [x] PROJ-12: Task title
		line := fmt.Sprintf("%s %s: %s", checkbox, taskID, title)

		// Add relation label if present (right-aligned)
		if relationLabel != "" {
			// Calculate padding to right-align the label
			lineLen := len(checkbox) + 1 + len(taskID) + 2 + len(title)
			paddingNeeded := width - lineLen - len(relationLabelPlain) - 8
			if paddingNeeded > 0 {
				line += strings.Repeat(" ", paddingNeeded) + relationLabel
			} else {
				line += " " + relationLabel
			}
		}

		// Apply cursor styling
		if i == cursorIdx {
			content.WriteString(selectedStyle.Render("> "+line) + "\n")
		} else {
			content.WriteString(normalStyle.Render("  "+line) + "\n")
		}
	}

	// Show "no results" if filter matched nothing
	if len(filteredItems) == 0 && filterText != "" {
		content.WriteString(dimStyle.Render("  No tasks match \""+filterText+"\"") + "\n")
	}

	// Show "no tasks" if no items exist (and no filter)
	if len(filteredItems) == 0 && filterText == "" {
		content.WriteString(dimStyle.Render("  No other tasks in this project") + "\n")
	}

	// Help text
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("Enter: toggle  Esc: close") + "\n")

	return content.String()
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
