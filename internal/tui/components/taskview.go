package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

type TaskViewProps struct {
	Task        *models.TaskDetail
	ColumnName  string
	PopupWidth  int
	PopupHeight int
}

func RenderTaskView(props TaskViewProps) string {
	task := props.Task

	contentWidth := props.PopupWidth - 8

	leftColWidth := (contentWidth * 80) / 100
	rightColWidth := contentWidth - leftColWidth - 1

	leftContent := RenderDescription(DescriptionProps{
		Description: task.Description,
		Width:       leftColWidth,
	})

	rightColumn := RenderMetadataColumn(MetadataColumnProps{
		Task:       task,
		ColumnName: props.ColumnName,
		Width:      rightColWidth,
		HasBorder:  true,
	})

	var leftParts []string

	taskIDStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Highlight))
	leftParts = append(leftParts, taskIDStyle.Render(fmt.Sprintf("Task #%d", task.ID)))
	leftParts = append(leftParts, "")

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Highlight))
	leftParts = append(leftParts, titleStyle.Render(task.Title))
	leftParts = append(leftParts, "")

	leftParts = append(leftParts, leftContent)

	// Add parent tasks section (if any)
	if len(task.ParentTasks) > 0 {
		leftParts = append(leftParts, "")
		leftParts = append(leftParts, RenderSubtaskSection("Parent Issues:", task.ParentTasks))
	}

	// Add child tasks section (if any)
	if len(task.ChildTasks) > 0 {
		leftParts = append(leftParts, "")
		leftParts = append(leftParts, RenderSubtaskSection("Child Issues:", task.ChildTasks))
	}

	leftParts = append(leftParts, "")
	leftParts = append(leftParts, "")
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
	leftParts = append(leftParts, footerStyle.Render("[l] labels  [p] parents  [c] children  [Esc/Space] close"))

	leftFullContent := strings.Join(leftParts, "\n")

	leftColumnFull := lipgloss.NewStyle().
		Width(leftColWidth).
		Padding(0, 1).
		Render(leftFullContent)

	rightColumnFull := rightColumn

	fullContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumnFull,
		rightColumnFull,
	)

	taskBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.Highlight)).
		Padding(1, 2).
		Width(props.PopupWidth).
		Height(props.PopupHeight).
		Render(fullContent)

	return taskBox
}

// RenderSubtaskRow renders a single subtask reference in the format "PROJ-12: Fix the thing"
func RenderSubtaskRow(ref *models.TaskReference) string {
	identifier := fmt.Sprintf("%s-%d", ref.ProjectName, ref.TicketNumber)

	identifierStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Highlight))

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Normal))

	return identifierStyle.Render(identifier) + ": " + titleStyle.Render(ref.Title)
}

// RenderSubtaskSection renders a section of subtasks with a header
func RenderSubtaskSection(header string, tasks []*models.TaskReference) string {
	if len(tasks) == 0 {
		return ""
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Bold(true)

	var parts []string
	parts = append(parts, headerStyle.Render(header))
	for _, task := range tasks {
		parts = append(parts, RenderSubtaskRow(task))
	}

	return strings.Join(parts, "\n")
}
