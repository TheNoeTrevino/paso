package components

import (
	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/config/colors"
)

// RenderDatabaseTypeChip renders a database type as a colored chip
// Uses Edit color for postgres (blue) and Accent for sqlite (purple)
func RenderDatabaseTypeChip(dbType string, colorScheme colors.ColorScheme, backgroundColor string) string {
	var chipColor string
	var displayText string

	switch dbType {
	case "postgres", "postgresql":
		chipColor = colorScheme.Edit // Blue for PostgreSQL
		displayText = "postgres"
	case "sqlite":
		chipColor = colorScheme.Accent // Purple for SQLite
		displayText = "sqlite"
	default:
		chipColor = colorScheme.Subtle // Gray for unknown
		displayText = dbType
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(chipColor)).
		Background(lipgloss.Color(backgroundColor)).
		Render(displayText)
}
