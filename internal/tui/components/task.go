package components

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// RenderTask renders a single task as a card
//
//		┌──────────────────────────────┐
//		│ {Task Title}                 │
//		│ type | priority | blocked    │
//		│ [label1] [label2]...         │
//		└──────────────────────────────┘
//	 This has a fixed width and length
func RenderTask(task *models.TaskSummary, selected bool) string {
	var bg string
	if selected {
		bg = theme.SelectedBg
	} else {
		bg = theme.TaskBg
	}

	title := renderTaskSummaryTitle(task, bg)
	metadataLine := renderTaskSummaryMetadata(task, bg)
	labelChips := renderTaskCardLabels(task.Labels, bg)
	content := title + metadataLine + labelChips

	style := TaskStyle.
		BorderForeground(lipgloss.Color(theme.SelectedBorder)).
		BorderBackground(lipgloss.Color(bg)).
		Background(lipgloss.Color(bg))

	return style.Render(content)
}

func renderTaskSummaryTitle(task *models.TaskSummary, bg string) string {
	title := task.Title

	if len(title) >= taskTitleMaxLength {
		title = title[:taskTitleMaxLength] + EllipsisStyle.
			Background(lipgloss.Color(bg)).
			Render("...")
	}

	return TaskTitleStyle.Render(" " + title)
}

// renderTaskCardLabels renders the labels as chips, with their color as the background
func renderTaskCardLabels(labels []*models.Label, bg string) string {
	if len(labels) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Background(lipgloss.Color(bg)).Italic(true)
		return "\n " + emptyStyle.Render("no labels")
	}

	// Calculate total raw length of all labels joined by spaces
	rawNames := make([]string, len(labels))
	for i, label := range labels {
		rawNames[i] = label.Name
	}
	aggregated := strings.Join(rawNames, " ")

	spacer := lipgloss.NewStyle().Background(lipgloss.Color(bg)).Render(" ")

	// If fits, render all labels
	if len(aggregated) <= taskLabelsMaxLength {
		var chips []string
		for _, label := range labels {
			chips = append(chips, RenderLabelChip(label, bg))
		}
		return "\n " + strings.Join(chips, spacer)
	}

	// Otherwise, add labels until we'd exceed the limit, then append ellipsis
	var chips []string
	currentLength := 0
	ellipsis := "..."

	for i, label := range labels {
		nextLength := currentLength + len(label.Name)
		if i > 0 {
			nextLength += 1 // space separator
		}

		// Check if adding this label + ellipsis would exceed limit
		if nextLength+len(ellipsis) > taskLabelsMaxLength {
			break
		}

		chips = append(chips, RenderLabelChip(label, bg))
		currentLength = nextLength
	}

	result := strings.Join(chips, spacer)
	result += EllipsisStyle.Background(lipgloss.Color(bg)).Render(ellipsis)

	return "\n " + result
}

// renderTaskSummaryMetadata Renders type, priority and blocked on the same line, separated by │
func renderTaskSummaryMetadata(task *models.TaskSummary, bg string) string {
	var typeDisplay string
	var priorityDisplay string

	if task.TypeDescription != "" {
		typeDisplay = SubtleStyle.Render(task.TypeDescription)
	} else {
		typeDisplay = SubtleStyle.Italic(true).Render("no type")
	}

	priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(task.PriorityColor)).Background(lipgloss.Color(bg))
	if task.PriorityDescription != "" && task.PriorityColor != "" {
		priorityDisplay = priorityStyle.Render(task.PriorityDescription)
	} else {
		priorityStyle = priorityStyle.
			Foreground(lipgloss.Color(theme.Subtle)).
			Italic(true)
		priorityDisplay = priorityStyle.Render("no priority")
	}

	separatorStyle := SubtleStyle.Background(lipgloss.Color(bg))
	separator := separatorStyle.Render(" │ ")

	result := typeDisplay + separator + priorityDisplay

	if task.IsBlocked {
		blockedDisplay := BlockedStyle.
			Background(lipgloss.Color(bg)).
			Render("blocked")

		result += separator + blockedDisplay
	}

	return "\n " + result
}
