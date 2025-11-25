package components

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/thenoetrevino/paso/internal/models"
)

type TaskViewProps struct {
	Task         *models.TaskDetail
	DB           *sql.DB
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
		Task:      task,
		DB:        props.DB,
		Width:     rightColWidth,
		HasBorder: true,
	})

	var leftParts []string

	taskIDStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170"))
	leftParts = append(leftParts, taskIDStyle.Render(fmt.Sprintf("Task #%d", task.ID)))
	leftParts = append(leftParts, "")

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170"))
	leftParts = append(leftParts, titleStyle.Render(task.Title))
	leftParts = append(leftParts, "")

	leftParts = append(leftParts, leftContent)

	leftParts = append(leftParts, "")
	leftParts = append(leftParts, "")
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
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
		BorderForeground(lipgloss.Color("170")).
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
