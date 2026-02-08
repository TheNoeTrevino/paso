package components

import (
	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/text"
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
	metadataLine := renderTaskSummaryMetadata(task, bg, width)
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
func renderTaskSummaryTitle(task *models.TaskSummary, bg string, cardWidth int) string {
	maxWidth := cardWidth - taskCardSafetyBuffer

	segment := text.NewStyledSegment(task.Title, TaskTitleStyle, bg)
	return " " + text.TruncateSegments([]text.Segment{segment}, "", maxWidth, bg)
}

// renderTaskCardLabels renders labels as colored chips, truncating with "..." if needed.
func renderTaskCardLabels(labels []*models.Label, bg string, cardWidth int) string {
	maxWidth := cardWidth - taskCardSafetyBuffer

	if len(labels) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Subtle)).
			Background(lipgloss.Color(bg)).
			Italic(true)
		return "\n " + emptyStyle.Render("no labels")
	}

	spacer := " "

	segments := make([]text.Segment, len(labels))
	for i, label := range labels {
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(label.Color))
		segments[i] = text.NewStyledSegment(label.Name, labelStyle, bg)
	}

	return "\n " + text.TruncateSegments(segments, spacer, maxWidth, bg)
}

// renderTaskSummaryMetadata renders type, priority, assignee and blocked on the same line, separated by |.
// Truncates with "..." if the total width exceeds the card width.
func renderTaskSummaryMetadata(task *models.TaskSummary, bg string, cardWidth int) string {
	maxWidth := cardWidth - taskCardSafetyBuffer

	var segments []text.Segment

	// Type
	if task.TypeDescription != "" {
		segments = append(segments, text.NewStyledSegment(task.TypeDescription, SubtleStyle, bg))
	} else {
		segments = append(segments, text.NewStyledSegment("no type", SubtleStyle.Italic(true), bg))
	}

	// Priority
	if task.PriorityDescription != "" && task.PriorityColor != "" {
		prioStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(task.PriorityColor))
		segments = append(segments, text.NewStyledSegment(task.PriorityDescription, prioStyle, bg))
	} else {
		noPrioStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Subtle)).
			Italic(true)
		segments = append(segments, text.NewStyledSegment("no priority", noPrioStyle, bg))
	}

	// Assignee
	if task.AssigneeName != nil {
		assigneeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight))
		segments = append(segments, text.NewStyledSegment("@"+*task.AssigneeName, assigneeStyle, bg))
	}

	// Blocked badge
	if task.IsBlocked {
		segments = append(segments, text.NewStyledSegment("blocked", BlockedStyle, bg))
	}

	separator := " │ "

	return "\n " + text.TruncateSegments(segments, separator, maxWidth, bg)
}
