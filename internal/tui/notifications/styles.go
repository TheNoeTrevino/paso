package notifications

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
			foreground:       "39",
			background:       "17",
			borderForeground: "17",
		}
	case Warning:
		return style{
			icon:             "⚠",
			title:            "Warning",
			foreground:       "220",
			background:       "94",
			borderForeground: "94",
		}
	case Error:
		return style{
			icon:             "✕",
			title:            "Error",
			foreground:       "196",
			background:       "52",
			borderForeground: "52",
		}
	default:
		return style{
			icon:             "🔔",
			title:            "Info",
			foreground:       "39",
			background:       "17",
			borderForeground: "17",
		}
	}
}
