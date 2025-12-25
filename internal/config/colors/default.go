package colors

// Default returns the default color scheme (purple theme)
func Default() *ColorScheme {
	return &ColorScheme{
		Preset: "default",

		// Primary
		Accent: "#874BFD",

		// Background
		Background:       "#1C1C1C",
		ColumnBackground: "#262626",

		// Semantic
		Create: "#5FD75F",
		Edit:   "#5F87D7",
		Delete: "#FF0000",

		// UI elements
		ColumnBorder:   "#5F87D7",
		TaskBorder:     "#585858",
		TaskBackground: "#262626",
		SelectedBorder: "#D75FD7",
		SelectedBg:     "#3A3A3A",

		// Text
		Title:  "#D75FD7",
		Subtle: "#585858",
		Normal: "#D0D0D0",

		// Notifications
		InfoFg:    "#00AFFF",
		InfoBg:    "#00005F",
		WarningFg: "#FFD700",
		WarningBg: "#875F00",
		ErrorFg:   "#FF0000",
		ErrorBg:   "#5F0000",

		// Status bar
		StatusBarBg:   "#874BFD", // Matches accent
		StatusBarText: "#D0D0D0", // Matches normal text
	}
}
