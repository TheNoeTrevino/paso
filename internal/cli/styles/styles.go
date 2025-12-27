package styles

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/models"
)

var (
	// Card styles
	CardStyle lipgloss.Style
	CardWidth = 80

	// Text styles
	TitleStyle    lipgloss.Style
	SubtitleStyle lipgloss.Style
	LabelStyle    lipgloss.Style // For field labels like "Type:", "Priority:"
	ValueStyle    lipgloss.Style // For field values
	SectionStyle  lipgloss.Style // For section headers like "Description", "Labels"

	// Status styles
	BlockedStyle lipgloss.Style
	SuccessStyle lipgloss.Style
	ErrorStyle   lipgloss.Style
	WarningStyle lipgloss.Style
)

// Init initializes all CLI styles with the given color scheme
func Init(colors config.ColorScheme) {
	CardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors.Accent)).
		Padding(1, 2).
		Width(CardWidth)

	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colors.Title))

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.Subtle))

	LabelStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colors.Accent))

	ValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.Normal))

	SectionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.Accent)).
		Bold(true).
		MarginTop(1)

	BlockedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colors.ErrorFg)).
		Background(lipgloss.Color(colors.ErrorBg)).
		Padding(0, 1)

	SuccessStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colors.InfoFg)).
		Background(lipgloss.Color(colors.InfoBg)).
		Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colors.ErrorFg)).
		Background(lipgloss.Color(colors.ErrorBg)).
		Padding(0, 1)

	WarningStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colors.WarningFg)).
		Background(lipgloss.Color(colors.WarningBg)).
		Padding(0, 1)
}

// ═══════════════════════════════════════════════════════════════════
// HELPER FUNCTIONS
// ═══════════════════════════════════════════════════════════════════

// ColoredText renders text with a hex color
func ColoredText(text, hexColor string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(hexColor)).
		Render(text)
}

// BoldColoredText renders bold text with a hex color
func BoldColoredText(text, hexColor string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(hexColor)).
		Render(text)
}

// RenderLabelChip renders a label as "[name]" with the label's color
func RenderLabelChip(label *models.Label) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(label.Color)).
		Bold(true).
		Render("[" + label.Name + "]")
}

// RenderTaskReference renders a task reference with colored bullet
// Format: "• ProjectName-123 - Title"
func RenderTaskReference(ref *models.TaskReference) string {
	bulletStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ref.RelationColor))

	taskRef := fmt.Sprintf("%s-%d - %s",
		ref.ProjectName,
		ref.TicketNumber,
		ref.Title)

	return bulletStyle.Render("• " + taskRef)
}

// RenderTaskReferenceWithLabel renders a task reference with relation label
// Format: "• ProjectName-123 - RelationLabel - Title"
func RenderTaskReferenceWithLabel(ref *models.TaskReference) string {
	bulletStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ref.RelationColor))

	taskRef := fmt.Sprintf("%s-%d - %s - %s",
		ref.ProjectName,
		ref.TicketNumber,
		ref.RelationLabel,
		ref.Title)

	return bulletStyle.Render("• " + taskRef)
}

// RenderCard wraps content in a styled card border
func RenderCard(content string) string {
	return CardStyle.Render(content)
}
