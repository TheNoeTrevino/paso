// Package components provides reusable UI components and styles.
// Call InitStyles() before use to initialize all style variables.
package components

import (
	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// These are cached to avoid recomputing on every redraw.
var (
	// compared to the defaults, these feel like
	// they take up less space
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
	TabStyle lipgloss.Style

	// ActiveTabStyle defines the selected tab
	ActiveTabStyle lipgloss.Style

	// TabGapStyle fills the remaining space after tabs
	TabGapStyle lipgloss.Style

	// ColumnStyle defines the appearance of kanban board columns
	ColumnStyle lipgloss.Style

	// TaskStyle defines the appearance of individual tasks as cards
	TaskStyle lipgloss.Style

	// ColumnTitleStyle defines the appearance of titles (column names, app header)
	ColumnTitleStyle lipgloss.Style

	// TaskTitleStyle defines the appearance of titles (column names, app header)
	TaskTitleStyle lipgloss.Style

	// FormBoxStyle defines the base style for ticket forms (purple border)
	FormBoxStyle lipgloss.Style

	// ProjectFormBoxStyle defines the base style for project creation forms (green border)
	ProjectFormBoxStyle lipgloss.Style

	// CreateInputBoxStyle defines the base style for creation dialogs (green border)
	CreateInputBoxStyle lipgloss.Style

	// EditInputBoxStyle defines the base style for edit dialogs (blue border)
	EditInputBoxStyle lipgloss.Style

	// EllipsisStyle defines the style for ellipses in truncated text
	EllipsisStyle lipgloss.Style

	// DeleteConfirmBoxStyle defines the base style for deletion confirmations (red border)
	DeleteConfirmBoxStyle lipgloss.Style

	// HelpBoxStyle defines the base style for help screen (blue border)
	HelpBoxStyle lipgloss.Style

	// LabelPickerBoxStyle defines the base style for label picker (purple border)
	LabelPickerBoxStyle lipgloss.Style

	// LabelPickerCreateBoxStyle defines the style for label creation mode (green border)
	LabelPickerCreateBoxStyle lipgloss.Style

	// InfoBannerStyle defines the appearance of info notifications (blue)
	InfoBannerStyle lipgloss.Style

	// WarningBannerStyle defines the appearance of warning notifications (yellow)
	WarningBannerStyle lipgloss.Style

	// ErrorBannerStyle defines the appearance of error messages (red)
	ErrorBannerStyle lipgloss.Style

	// IndicatorStyle defines the appearance of scroll indicators
	IndicatorStyle lipgloss.Style

	// IndicatorStyle defines the appearance of scroll indicators
	BlockedStyle lipgloss.Style

	// StatusBarStyle defines the base style for the status bar
	StatusBarStyle lipgloss.Style

	// StatusBarSearchStyle defines the style for the search section in the status bar
	StatusBarSearchStyle lipgloss.Style

	// SubtleStyle defines a base style for subtle text
	SubtleStyle lipgloss.Style
)

// InitStyles initializes all styles with the given color scheme
func InitStyles(colors colors.ColorScheme) {
	theme.Init(colors)

	TabStyle = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(lipgloss.Color(theme.Highlight)).
		Padding(0, 1)

	ActiveTabStyle = TabStyle.Border(activeTabBorder, true)

	TabGapStyle = TabStyle.
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	ColumnStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors.ColumnBorder)).
		PaddingLeft(1).
		PaddingRight(1).
		Width(40)

	TaskStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color(colors.TaskBorder)).
		BorderBackground(lipgloss.Color(colors.TaskBackground)).
		Background(lipgloss.Color(colors.TaskBackground)).
		Padding(0).
		Width(36)

	TaskTitleStyle = lipgloss.NewStyle().
		Bold(true)

	ColumnTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colors.Title))

	FormBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors.Accent)).
		Padding(1, 2)

	ProjectFormBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors.Create)).
		Padding(1, 2)

	CreateInputBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors.Create)).
		Padding(1)

	EditInputBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors.Edit)).
		Padding(1)

	DeleteConfirmBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors.Delete)).
		Padding(1)

	HelpBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors.Edit)).
		Padding(1, 2)

	LabelPickerBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors.Accent)).
		Padding(1, 2)

	LabelPickerCreateBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors.Create)).
		Padding(1, 2)

	InfoBannerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.InfoFg)).
		Background(lipgloss.Color(colors.InfoBg)).
		Bold(true).
		Padding(0, 1)

	WarningBannerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.WarningFg)).
		Background(lipgloss.Color(colors.WarningBg)).
		Bold(true).
		Padding(0, 1)

	ErrorBannerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.ErrorFg)).
		Background(lipgloss.Color(colors.ErrorBg)).
		Bold(true).
		Padding(0, 1)

	IndicatorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Align(lipgloss.Center)

	StatusBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(colors.StatusBarBg)).
		Foreground(lipgloss.Color(colors.StatusBarText))

	StatusBarSearchStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(colors.Background))

	EllipsisStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle))

	BlockedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ErrorFg))

	SubtleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle))
}
