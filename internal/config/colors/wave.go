package colors

// Wave returns the Kanagawa Wave color scheme (dark theme with blue/purple accents)
func Wave() *ColorScheme {
	return &ColorScheme{
		Preset: "wave",

		// Primary accent color
		Accent: palette.oniViolet,

		// Background colors
		Background:       palette.sumiInk1,
		ColumnBackground: palette.sumiInk2,

		// Semantic colors
		Create: palette.springGreen,
		Edit:   palette.crystalBlue,
		Delete: palette.peachRed,

		// UI element colors
		ColumnBorder:   palette.sumiInk6,
		TaskBorder:     palette.sumiInk4,
		TaskBackground: palette.sumiInk3,
		SelectedBorder: palette.waveAqua2,
		SelectedBg:     palette.waveBlue1,

		// Text colors
		Title:  palette.crystalBlue,
		Subtle: palette.fujiGray,
		Normal: palette.fujiWhite,

		// Notification colors
		InfoFg:    palette.dragonBlue,
		InfoBg:    palette.winterBlue,
		WarningFg: palette.roninYellow,
		WarningBg: palette.winterYellow,
		ErrorFg:   palette.samuraiRed,
		ErrorBg:   palette.winterRed,
	}
}
