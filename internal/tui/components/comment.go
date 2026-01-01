package components

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// RenderCommentCard renders a single comment as a card
//
//	┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  (selected)
//	┃ 󰀄 noetest    Dec 28 18:54  edited    ┃
//	┃ this is a new comment                ┃
//	┃ with multiple lines                  ┃
//	┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
//
//	┌──────────────────────────────────────────┐  (unselected)
//	│ 󰀄 noetest    Dec 28 17:32                │
//	│ another comment here                     │
//	└──────────────────────────────────────────┘
func RenderCommentCard(comment *models.Comment, selected bool, width int) string {
	// Background color based on selection (same as task cards)
	var bg string
	if selected {
		bg = theme.SelectedBg
	} else {
		bg = theme.TaskBg
	}

	// Render header and content
	header := renderCommentHeader(comment)
	content := renderCommentContent(comment, width, bg)

	// Combine header and content
	fullContent := header + "\n" + content

	// Border style - same pattern as task cards
	var borderForeground string
	if selected {
		borderForeground = theme.SelectedBorder
	} else {
		borderForeground = theme.Subtle
	}

	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color(borderForeground)).
		BorderBackground(lipgloss.Color(bg)).
		Background(lipgloss.Color(bg)).
		Width(width).
		Padding(0, 1)

	return style.Render(fullContent)
}

// renderCommentHeader renders the comment header with author, date, and edit indicator
// Format: 󰀄 {author}  {created_date}  (edited {updated_date})
func renderCommentHeader(comment *models.Comment) string {
	// Author icon + author name
	authorIcon := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Render("󰀄 ")
	author := comment.Author

	// Date icon + created date
	dateIcon := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Render("  ")
	createdDate := comment.CreatedAt.Format("Jan 2 15:04")

	// Edited indicator (only if updated_at > created_at)
	var editedIndicator string
	if !comment.UpdatedAt.IsZero() && comment.UpdatedAt.After(comment.CreatedAt) {
		editedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Subtle)).
			Italic(true)

		editedDate := comment.UpdatedAt.Format("Jan 2 15:04")
		editedIndicator = " " + editedStyle.
			Render(fmt.Sprintf("(edited %s)", editedDate))
	}

	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Normal))
	return headerStyle.Render(authorIcon + author + dateIcon + createdDate + editedIndicator)
}

// renderCommentContent renders the comment content with word wrapping
func renderCommentContent(comment *models.Comment, width int, bg string) string {
	// Reserve space for padding/borders
	contentWidth := max(width-4, 20)

	// Wrap content to width
	wrapped := wordwrap.String(comment.Message, contentWidth)

	// Apply styling with background
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Normal)).
		Background(lipgloss.Color(bg))

	return contentStyle.Render(wrapped)
}
