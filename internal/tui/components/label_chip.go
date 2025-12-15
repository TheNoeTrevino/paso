package components

import (
	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
)

// RenderLabelChip renders a single label as a small colored chip
func RenderLabelChip(label *models.Label, backgroundColor string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(label.Color)).
		Background(lipgloss.Color(backgroundColor)).
		Render(label.Name)
}
