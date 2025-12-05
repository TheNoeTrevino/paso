package notifications

import "github.com/thenoetrevino/paso/internal/tui/theme"

type style struct {
	icon             string
	title            string
	foreground       string
	background       string
	borderForeground string
}

func (s Severity) style() style {
	switch s {
	case Info:
		return style{
			icon:             "🔔",
			title:            "Info",
			foreground:       theme.InfoFg,
			background:       theme.InfoBg,
			borderForeground: theme.InfoBg,
		}
	case Warning:
		return style{
			icon:             "⚠",
			title:            "Warning",
			foreground:       theme.WarningFg,
			background:       theme.WarningBg,
			borderForeground: theme.WarningBg,
		}
	case Error:
		return style{
			icon:             "✕",
			title:            "Error",
			foreground:       theme.ErrorFg,
			background:       theme.ErrorBg,
			borderForeground: theme.ErrorBg,
		}
	default:
		return style{
			icon:             "🔔",
			title:            "Info",
			foreground:       theme.InfoFg,
			background:       theme.InfoBg,
			borderForeground: theme.InfoBg,
		}
	}
}
