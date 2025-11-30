package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/thenoetrevino/paso/internal/models"
)

type MetadataColumnProps struct {
	Task       *models.TaskDetail
	ColumnName string
	Width      int
	HasBorder  bool
}

func RenderMetadataColumn(props MetadataColumnProps) string {
	var parts []string

	labelHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Bold(true)

	// Status section
	parts = append(parts, renderMetadataSection("Status", props.ColumnName))

	// Created timestamp
	createdStr := props.Task.CreatedAt.Format("Jan 2, 2006 3:04 PM")
	parts = append(parts, renderMetadataSection("Created", createdStr))

	// Updated timestamp
	updatedStr := props.Task.UpdatedAt.Format("Jan 2, 2006 3:04 PM")
	parts = append(parts, renderMetadataSection("Updated", updatedStr))

	// Labels section
	if len(props.Task.Labels) > 0 {
		parts = append(parts, labelHeaderStyle.Render("Labels"))
		for _, label := range props.Task.Labels {
			parts = append(parts, renderLabelChip(label))
		}
		parts = append(parts, "")
	} else {
		parts = append(parts, labelHeaderStyle.Render("Labels"))
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		parts = append(parts, emptyStyle.Render("No labels"))
		parts = append(parts, "")
	}

	content := strings.Join(parts, "\n")

	style := lipgloss.NewStyle().
		Width(props.Width).
		Padding(0, 1)

	if props.HasBorder {
		style = style.
			BorderLeft(true).
			BorderStyle(lipgloss.Border{
				Left: "│",
			}).
			BorderForeground(lipgloss.Color("240"))
	}

	return style.Render(content)
}

func renderMetadataSection(label string, value string) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	return labelStyle.Render(label) + "\n" + valueStyle.Render(value) + "\n"
}

func renderLabelChip(label *models.Label) string {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(label.Color)).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		MarginRight(1).
		Render(label.Name)
}
