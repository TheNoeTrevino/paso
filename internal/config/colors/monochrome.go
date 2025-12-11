package colors

// Monochrome returns a black and white color scheme
func Monochrome() *ColorScheme {
	return &ColorScheme{
		Preset: "monochrome",

		// Primary
		Accent: "#FFFFFF",

		// Background
		Background:       "#121212",
		ColumnBackground: "#1C1C1C",

		// Semantic
		Create: "#FFFFFF",
		Edit:   "#FFFFFF",
		Delete: "#FFFFFF",

		// UI elements
		ColumnBorder:   "#FFFFFF",
		TaskBorder:     "#585858",
		TaskBackground: "#1C1C1C",
		SelectedBorder: "#FFFFFF",
		SelectedBg:     "#3A3A3A",

		// Text
		Title:  "#FFFFFF",
		Subtle: "#585858",
		Normal: "#D0D0D0",

		// Notifications
		InfoFg:    "#FFFFFF",
		InfoBg:    "#1C1C1C",
		WarningFg: "#FFFFFF",
		WarningBg: "#3A3A3A",
		ErrorFg:   "#FFFFFF",
		ErrorBg:   "#585858",
	}
}
