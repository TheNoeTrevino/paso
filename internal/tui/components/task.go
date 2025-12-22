package components

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// RenderTask renders a single task as a card
// This is a pure, reusable component that displays task title and labels
//
// Format (as a card with border):
//
//	┌─────────────────────┐
//	│ {Task Title}        │
//	│ type | priority     │
//	│ [label1] [label2]   │
//	└─────────────────────┘
//
// All three content lines are ALWAYS displayed to maintain consistent card height.
// When selected is true, the task is highlighted with:
//   - Bold text
//   - Purple border color
//   - Brighter background
func RenderTask(task *models.TaskSummary, selected bool) string {
	var bg string
	if selected {
		bg = theme.SelectedBg
	} else {
		bg = theme.TaskBg
	}

	// Add blocked indicator if task is blocked
	var blockedIndicator string
	if task.IsBlocked {
		blockedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true)
		blockedIndicator = blockedStyle.Render("! ")
	}

	// Format task content with title (add leading space for padding)
	title := lipgloss.NewStyle().Bold(true).Render(" 󰗴 " + task.Title + blockedIndicator)

	// Render type and priority on the same line, separated by │
	var typeDisplay string
	var priorityDisplay string

	// Type display - always show, use placeholder if missing
	if task.TypeDescription != "" {
		typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
		typeDisplay = typeStyle.Render(task.TypeDescription)
	} else {
		typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Italic(true)
		typeDisplay = typeStyle.Render("no type")
	}

	// Priority display with color - always show, use placeholder if missing
	if task.PriorityDescription != "" && task.PriorityColor != "" {
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(task.PriorityColor)).Background(lipgloss.Color(bg))
		priorityDisplay = priorityStyle.Render(task.PriorityDescription)
	} else {
		// Default placeholder if not set
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Background(lipgloss.Color(bg)).Italic(true)
		priorityDisplay = priorityStyle.Render("no priority")
	}

	// Separator
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Background(lipgloss.Color(bg))
	separator := separatorStyle.Render(" │ ")

	// Combine type and priority - always include this line
	metadataLine := "\n " + typeDisplay + separator + priorityDisplay

	// Render label chips - ALWAYS include this line even if empty to maintain fixed height
	labelChips := renderTaskCardLabels(task.Labels, bg)

	content := title + metadataLine + labelChips

	style := TaskStyle.
		BorderForeground(lipgloss.Color(theme.SelectedBorder)).
		BorderBackground(lipgloss.Color(bg)).
		Background(lipgloss.Color(bg))

	return style.Render(content)
}

func renderTaskCardLabels(labels []*models.Label, bg string) string {
	spacer := lipgloss.NewStyle().Background(lipgloss.Color(bg)).Render(" ")
	var labelChips string
	if len(labels) > 0 {
		var chips []string
		for _, label := range labels {
			chips = append(chips, RenderLabelChip(label, bg))
		}
		labelChips = "\n " + strings.Join(chips, spacer)
	} else {
		// place holder for no labels
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Background(lipgloss.Color(bg)).Italic(true)
		labelChips = "\n " + emptyStyle.Render("no labels")
	}
	return labelChips
}
