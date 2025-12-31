package components

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// RenderTask renders a single task as a card
//
//		┌─────────────────────┐
//		│ {Task Title}        │
//		│ type | priority     │
//		│ [label1] [label2]   │
//		└─────────────────────┘
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
	var blockedIndicator string
	if task.IsBlocked {
		blockedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true).Italic(true)
		blockedIndicator = blockedStyle.Render("! ") // FIXME: why is this sometimes right and sometimes wrong
	}
	title := task.Title

	if len(title) >= 30 {
		ellipsisStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Subtle)).
			Background(lipgloss.Color(bg)).
			Italic(true)
		title = title[:27] + ellipsisStyle.Render("...")
	}

	if len(title) < 30 {
		title = title + strings.Repeat(" ", 30-len(title))
	}

	return lipgloss.NewStyle().
		Bold(true).
		Render(" 󰗴 " + blockedIndicator + title)
}

// renderTaskCardLabels renders the labels as chips, with their color as the background
func renderTaskCardLabels(labels []*models.Label, bg string) string {
	if len(labels) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Background(lipgloss.Color(bg)).Italic(true)
		return "\n " + emptyStyle.Render("no labels")
	}

	spacer := lipgloss.NewStyle().Background(lipgloss.Color(bg)).Render(" ")
	var chips []string
	for _, label := range labels {
		chips = append(chips, RenderLabelChip(label, bg))
	}
	labelChips := strings.Join(chips, spacer)
	return "\n " + labelChips
}

// renderTaskSummaryMetadata Renders type and priority on the same line, separated by │
func renderTaskSummaryMetadata(task *models.TaskSummary, bg string) string {
	var typeDisplay string
	var priorityDisplay string

	if task.TypeDescription != "" {
		typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
		typeDisplay = typeStyle.Render(task.TypeDescription)
	} else {
		typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Italic(true)
		typeDisplay = typeStyle.Render("no type")
	}

	if task.PriorityDescription != "" && task.PriorityColor != "" {
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(task.PriorityColor)).Background(lipgloss.Color(bg))
		priorityDisplay = priorityStyle.Render(task.PriorityDescription)
	} else {
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Background(lipgloss.Color(bg)).Italic(true)
		priorityDisplay = priorityStyle.Render("no priority")
	}

	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Background(lipgloss.Color(bg))
	separator := separatorStyle.Render(" │ ")

	return "\n " + typeDisplay + separator + priorityDisplay
}
