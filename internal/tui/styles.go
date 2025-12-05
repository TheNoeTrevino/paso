package tui

import "charm.land/lipgloss/v2"

// Style definitions for the kanban board UI
// These styles follow Lipgloss conventions for composable terminal styling

var (
	// Tab colors
	highlight = lipgloss.Color("#874BFD")
	subtle    = lipgloss.Color("#D9DCCF")

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

	// Modal dialog base styles (extract width/height for dynamic sizing)

	// FormBoxStyle defines the base style for ticket forms (purple border)
	FormBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("170")).
			Padding(1, 2)

	// ProjectFormBoxStyle defines the base style for project creation forms (green border)
	ProjectFormBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("42")).
				Padding(1, 2)

	// CreateInputBoxStyle defines the base style for creation dialogs (green border)
	CreateInputBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("42")).
				Padding(1)

	// EditInputBoxStyle defines the base style for edit dialogs (blue border)
	EditInputBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(1)

	// DeleteConfirmBoxStyle defines the base style for deletion confirmations (red border)
	DeleteConfirmBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("196")).
				Padding(1)

	// HelpBoxStyle defines the base style for help screen (blue border)
	HelpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	// LabelPickerBoxStyle defines the base style for label picker (purple border)
	LabelPickerBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("170")).
				Padding(1, 2)

	// LabelPickerCreateBoxStyle defines the style for label creation mode (green border)
	LabelPickerCreateBoxStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("42")).
					Padding(1, 2)

	// InfoBannerStyle defines the appearance of info notifications (blue)
	InfoBannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")). // Bright blue
			Background(lipgloss.Color("17")). // Dark blue
			Bold(true).
			Padding(0, 1)

	// WarningBannerStyle defines the appearance of warning notifications (yellow)
	WarningBannerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("220")). // Yellow
				Background(lipgloss.Color("94")).  // Dark yellow/brown
				Bold(true).
				Padding(0, 1)

	// ErrorBannerStyle defines the appearance of error messages (red)
	ErrorBannerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")). // Bright red
				Background(lipgloss.Color("52")).  // Dark red
				Bold(true).
				Padding(0, 1)
)
