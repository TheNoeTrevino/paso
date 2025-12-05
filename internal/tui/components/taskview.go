package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

type TaskViewProps struct {
	Task         *models.TaskDetail
	ColumnName   string
	PopupWidth   int
	PopupHeight  int
	ScreenWidth  int
	ScreenHeight int
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

	leftParts = append(leftParts, "")
	leftParts = append(leftParts, "")
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
	leftParts = append(leftParts, footerStyle.Render("[l] labels  [Esc/Space] close"))

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

	return lipgloss.Place(
		props.ScreenWidth, props.ScreenHeight,
		lipgloss.Center, lipgloss.Center,
		taskBox,
	)
}
