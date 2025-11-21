package tui

import "github.com/charmbracelet/lipgloss"

// Style definitions for the kanban board UI
// These styles follow Lipgloss conventions for composable terminal styling

var (
	// Tab colors
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}

	// Tab borders - active tab has no bottom border to "open" into content
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	// TabStyle defines inactive tabs
	TabStyle = lipgloss.NewStyle().
			Border(tabBorder, true).
			BorderForeground(highlight).
			Padding(0, 1)

	// ActiveTabStyle defines the selected tab
	ActiveTabStyle = TabStyle.Border(activeTabBorder, true)

	// TabGapStyle fills the remaining space after tabs
	TabGapStyle = TabStyle.
			BorderTop(false).
			BorderLeft(false).
			BorderRight(false)

	// ColumnStyle defines the appearance of kanban board columns
	ColumnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")). // Blue border
			Padding(1).
			Width(40) // Wider columns for better readability

	// TaskStyle defines the appearance of individual tasks as cards
	TaskStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")). // Gray border
			Background(lipgloss.Color("235")).       // Dark gray background
			Padding(1).
			MarginBottom(1).
			Width(36) // Slightly narrower than column to fit with padding

	// TitleStyle defines the appearance of titles (column names, app header)
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")) // Purple text
)
