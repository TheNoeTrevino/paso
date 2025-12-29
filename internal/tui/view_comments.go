package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// renderCommentsViewContent renders the full comments view content (without layer wrapping)
func (m Model) renderCommentsViewContent(width, height int) string {
	if m.CommentState.TaskID == 0 {
		return "Error: No task selected"
	}

	// Get task title for header
	task := m.getCurrentTask()
	taskTitle := "Unknown Task"
	if task != nil {
		taskTitle = task.Title
	}

	// Title bar
	commentCount := len(m.CommentState.Items)
	titleText := fmt.Sprintf("Task Comments - \"%s\" (%d comments)", taskTitle, commentCount)
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Highlight))
	titleBar := titleStyle.Render(titleText)

	// Empty state
	if m.CommentState.IsEmpty() {
		emptyContent := renderEmptyCommentsState(width, height-4) // Reserve for title + help
		helpText := renderCommentsHelpText()
		return lipgloss.JoinVertical(lipgloss.Left, titleBar, "", emptyContent, "", helpText)
	}

	// Calculate card dimensions (80% of width, responsive)
	cardWidth := (width * 8) / 10
	if cardWidth < 60 {
		cardWidth = 60
	}
	if cardWidth > 100 {
		cardWidth = 100
	}

	// Render comment cards
	var cards []string
	for i, item := range m.CommentState.Items {
		selected := (i == m.CommentState.Cursor)
		card := components.RenderCommentCard(item.Comment, selected, cardWidth)
		cards = append(cards, card)
	}

	// Join cards with 1 blank line spacing
	cardsContent := strings.Join(cards, "\n\n")

	// Calculate scroll indicators
	availableHeight := height - 4 // Reserve for title + help
	scrollIndicators := calculateCommentsScrollIndicators(m.CommentState, availableHeight, len(cards))

	// Combine content
	helpText := renderCommentsHelpText()

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleBar,
		"",
		scrollIndicators.Top,
		cardsContent,
		scrollIndicators.Bottom,
		"",
		helpText,
	)

	return content
}

// renderEmptyCommentsState renders the empty state when there are no comments
func renderEmptyCommentsState(width, height int) string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Italic(true).
		Align(lipgloss.Center).
		Width(width)

	lines := []string{
		"",
		"No comments yet.",
		"",
		"Press 'a' to add your first comment.",
		"",
		"Comments help you track context and reasoning",
		"as you work through tasks.",
		"",
	}

	return emptyStyle.Render(strings.Join(lines, "\n"))
}

// renderCommentsHelpText renders the help text at the bottom of the comments view
func renderCommentsHelpText() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle))

	return helpStyle.Render("[↑↓: navigate | Enter/e: edit | a: add | d: delete | Esc: close]")
}

// ScrollIndicators holds the top and bottom scroll indicator strings
type ScrollIndicators struct {
	Top    string
	Bottom string
}

// calculateCommentsScrollIndicators determines if scroll indicators should be shown
func calculateCommentsScrollIndicators(commentState *state.CommentState, availableHeight int, cardCount int) ScrollIndicators {
	indicators := ScrollIndicators{
		Top:    "",
		Bottom: "",
	}

	// For now, simple logic: if we have more than a few cards, show indicators
	// TODO: Implement proper viewport scrolling logic
	if cardCount > 3 {
		indicatorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Subtle)).
			Align(lipgloss.Center)

		if commentState.Cursor > 0 {
			indicators.Top = indicatorStyle.Render("↑ More comments above...")
		}

		if commentState.Cursor < cardCount-1 {
			indicators.Bottom = indicatorStyle.Render("↓ More comments below...")
		}
	}

	return indicators
}
