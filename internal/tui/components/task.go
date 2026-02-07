package components

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// RenderTask renders a single task as a card with dynamic width
//
//	┌──────────────────────────────┐
//	│ {Task Title}                 │
//	│ type | priority | blocked    │
//	│ [label1] [label2]...         │
//	└──────────────────────────────┘
//
// Parameters:
//   - task: The task summary to render
//   - selected: Whether this task is currently selected
//   - width: The width of the task card content area
func RenderTask(task *models.TaskSummary, selected bool, width int) string {
	var bg string
	if selected {
		bg = theme.SelectedBg
	} else {
		bg = theme.TaskBg
	}

	title := renderTaskSummaryTitle(task, bg, width)
	metadataLine := renderTaskSummaryMetadata(task, bg)
	labelChips := renderTaskCardLabels(task.Labels, bg, width)
	content := title + metadataLine + labelChips

	style := TaskStyle.
		Width(width).
		BorderForeground(lipgloss.Color(theme.SelectedBorder)).
		BorderBackground(lipgloss.Color(bg)).
		Background(lipgloss.Color(bg))

	return style.Render(content)
}

// renderTaskSummaryTitle renders the task title, truncating with "..." if needed.
// Uses lipgloss.Width() for measurement to match Lipgloss's word-wrap behavior.
func renderTaskSummaryTitle(task *models.TaskSummary, bg string, cardWidth int) string {
	// Safety buffer accounts for measurement discrepancies between
	// lipgloss.Width() and Lipgloss's internal word-wrap calculations
	const safetyBuffer = 2
	maxWidth := cardWidth - safetyBuffer

	title := task.Title
	styledEllipsis := EllipsisStyle.
		Background(lipgloss.Color(bg)).
		Render("...")

	// Check if title fits without truncation
	candidate := TaskTitleStyle.Render(" " + title)
	if lipgloss.Width(candidate) <= maxWidth {
		return candidate
	}

	// Truncate progressively until the rendered output fits
	runes := []rune(title)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		truncated := string(runes)
		candidate = TaskTitleStyle.Render(" " + truncated + styledEllipsis)
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
	}

	return TaskTitleStyle.Render(" " + styledEllipsis)
}

// renderTaskCardLabels renders labels as colored chips, truncating with "..." if needed.
// Uses lipgloss.Width() for measurement to match Lipgloss's word-wrap behavior.
func renderTaskCardLabels(labels []*models.Label, bg string, cardWidth int) string {
	const leadingSpace = 1
	const ellipsisWidth = 3

	maxWidth := cardWidth - leadingSpace

	if len(labels) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Subtle)).
			Background(lipgloss.Color(bg)).
			Italic(true)
		return "\n " + emptyStyle.Render("no labels")
	}

	spacer := lipgloss.NewStyle().Background(lipgloss.Color(bg)).Render(" ")

	// Calculate total width of all labels joined by spaces
	totalWidth := 0
	for i, label := range labels {
		if i > 0 {
			totalWidth += 1 // space separator
		}
		totalWidth += lipgloss.Width(label.Name)
	}

	// If all labels fit, render them all
	if totalWidth <= maxWidth {
		var chips []string
		for _, label := range labels {
			chips = append(chips, RenderLabelChip(label, bg))
		}
		return "\n " + strings.Join(chips, spacer)
	}

	// Otherwise, add labels until we'd exceed the limit, then append ellipsis
	var chips []string
	currentWidth := 0

	for i, label := range labels {
		labelWidth := lipgloss.Width(label.Name)
		nextWidth := currentWidth + labelWidth
		if i > 0 {
			nextWidth += 1 // space separator
		}

		// Check if adding this label + ellipsis would exceed limit
		if nextWidth+ellipsisWidth > maxWidth {
			break
		}

		chips = append(chips, RenderLabelChip(label, bg))
		currentWidth = nextWidth
	}

	result := strings.Join(chips, spacer)
	result += EllipsisStyle.Background(lipgloss.Color(bg)).Render("...")

	return "\n " + result
}

// renderTaskSummaryMetadata Renders type, priority, assignee and blocked on the same line, separated by │
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

	if task.AssigneeName != nil {
		assigneeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Highlight)).
			Background(lipgloss.Color(bg))
		result += separator + assigneeStyle.Render("@"+*task.AssigneeName)
	}

	if task.IsBlocked {
		blockedDisplay := BlockedStyle.
			Background(lipgloss.Color(bg)).
			Render("blocked")

		result += separator + blockedDisplay
	}

	return "\n " + result
}
