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
			icon:             "ⓘ",
			title:            "Info",
			foreground:       theme.InfoFg,
			background:       theme.Background,
			borderForeground: theme.InfoFg,
		}
	case Warning:
		return style{
			icon:             "⚠",
			title:            "Warning",
			foreground:       theme.WarningFg,
			background:       theme.Background,
			borderForeground: theme.WarningFg,
		}
	case Error:
		return style{
			icon:             "✕",
			title:            "Error",
			foreground:       theme.ErrorFg,
			background:       theme.Background,
			borderForeground: theme.ErrorFg,
		}
	default:
		return style{
			icon:             "ⓘ",
			title:            "Info",
			foreground:       theme.InfoFg,
			background:       theme.Background,
			borderForeground: theme.InfoFg,
		}
	}
}
