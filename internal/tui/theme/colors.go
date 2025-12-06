package theme

import "github.com/thenoetrevino/paso/internal/config"

// Colors holds the current theme colors, initialized by Init
var (
	Highlight      string
	Subtle         string
	Normal         string
	Create         string
	SelectedBorder string
	SelectedBg     string
	TaskBg         string
	InfoFg         string
	InfoBg         string
	WarningFg      string
	WarningBg      string
	ErrorFg        string
	ErrorBg        string
)

// Init initializes the theme colors from the given color scheme
func Init(colors config.ColorScheme) {
	Highlight = colors.Accent
	Subtle = colors.Subtle
	Normal = colors.Normal
	Create = colors.Create
	SelectedBorder = colors.SelectedBorder
	SelectedBg = colors.SelectedBg
	TaskBg = colors.TaskBackground
	InfoFg = colors.InfoFg
	InfoBg = colors.InfoBg
	WarningFg = colors.WarningFg
	WarningBg = colors.WarningBg
	ErrorFg = colors.ErrorFg
	ErrorBg = colors.ErrorBg
}
