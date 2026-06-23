// Package styles defines and initializes CLI styles using lipgloss
package styles

import (
	"fmt"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

// CardPaddingX is the horizontal padding applied inside the card border
// (on each side). CardBorderX is the width the rounded border consumes on each
// side. Both are subtracted from CardWidth to get the usable text width, since
// lipgloss counts the border within the configured Width.
const (
	CardPaddingX = 2
	CardBorderX  = 1
)

var (
	once sync.Once

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
func Init(colors colors.ColorScheme) {
	once.Do(func() {
		CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colors.Accent)).
			Padding(1, CardPaddingX).
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
	})
}

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

// RenderTaskReference renders a task reference as an indented, word-wrapped
// bullet. Format: "  • ProjectName-123 - Title". Long titles wrap with a
// hanging indent so continuation lines stay aligned under the bullet text.
func RenderTaskReference(ref *models.TaskReference) string {
	taskRef := fmt.Sprintf("%s-%d - %s",
		ref.ProjectName,
		ref.TaskNumber,
		ref.Title)

	return renderBullet(taskRef, ref.RelationColor)
}

// RenderTaskReferenceWithLabel renders a task reference with relation label
// Format: "  • ProjectName-123 - RelationLabel - Title".
func RenderTaskReferenceWithLabel(ref *models.TaskReference) string {
	taskRef := fmt.Sprintf("%s-%d - %s - %s",
		ref.ProjectName,
		ref.TaskNumber,
		ref.RelationLabel,
		ref.Title)

	return renderBullet(taskRef, ref.RelationColor)
}

// renderBullet renders text as a colored bullet list item, word-wrapped to the
// card width. The first line is indented two spaces (under the section body)
// and any wrapped continuation lines are indented four spaces so they align
// under the bullet's text rather than collapsing to the card margin.
func renderBullet(text, color string) string {
	const firstIndent = "  "
	const hangIndent = "    "

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	lines := WrapText("• "+text, CardContentWidth()-len(hangIndent))

	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			b.WriteString("\n")
		}
		indent := firstIndent
		if i > 0 {
			indent = hangIndent
		}
		b.WriteString(indent + style.Render(line))
	}
	return b.String()
}

// CardContentWidth returns the usable text width inside a card, i.e. the card
// width minus its border and horizontal padding on both sides.
func CardContentWidth() int {
	return CardWidth - (CardBorderX+CardPaddingX)*2
}

// WrapText word-wraps s to at most limit cells per line, breaking on word
// boundaries and hard-breaking any single word longer than the limit. Existing
// newlines in s are preserved as hard line breaks.
func WrapText(s string, limit int) []string {
	if limit < 1 {
		limit = 1
	}
	wrapped := wrap.String(wordwrap.String(s, limit), limit)
	return strings.Split(wrapped, "\n")
}

// RenderCard wraps content in a styled card border
func RenderCard(content string) string {
	return CardStyle.Render(content)
}
