package theme

import "github.com/thenoetrevino/paso/internal/config/colors"

// Colors holds the current theme colors, initialized by Init
var (
	Highlight      string
	Subtle         string
	Normal         string
	Create         string
	Blocked        string
	Background     string
	SelectedBorder string
	SelectedBg     string
	TaskBg         string
	ColumnBg       string
	InfoFg         string
	InfoBg         string
	WarningFg      string
	WarningBg      string
	ErrorFg        string
	ErrorBg        string
	StatusBarBg    string
	StatusBarText  string
)

// Init initializes the theme colors from the given color scheme
func Init(colors colors.ColorScheme) {
	Highlight = colors.Accent
	Subtle = colors.Subtle
	Normal = colors.Normal
	Create = colors.Create
	Blocked = colors.Blocked
	Background = colors.Background
	SelectedBorder = colors.SelectedBorder
	SelectedBg = colors.SelectedBg
	TaskBg = colors.TaskBackground
	ColumnBg = colors.ColumnBackground
	InfoFg = colors.InfoFg
	InfoBg = colors.InfoBg
	WarningFg = colors.WarningFg
	WarningBg = colors.WarningBg
	ErrorFg = colors.ErrorFg
	ErrorBg = colors.ErrorBg
	StatusBarBg = colors.StatusBarBg
	StatusBarText = colors.StatusBarText
}
