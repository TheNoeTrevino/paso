package colors

// Lotus returns the Kanagawa Lotus color scheme (light theme with cream/paper background)
func Lotus() *ColorScheme {
	return &ColorScheme{
		Preset: "lotus",

		// Primary accent color
		Accent: palette.lotusViolet4,

		// Background colors
		Background:       palette.lotusWhite0,
		ColumnBackground: palette.lotusWhite2,

		// Semantic colors
		Create: palette.lotusGreen,
		Edit:   palette.lotusBlue4,
		Delete: palette.lotusRed,

		// UI element colors
		ColumnBorder:   palette.lotusViolet1,
		TaskBorder:     palette.lotusWhite4,
		TaskBackground: palette.lotusWhite3,
		SelectedBorder: palette.lotusAqua,
		SelectedBg:     palette.lotusBlue1,

		// Text colors
		Title:  palette.lotusBlue4,
		Subtle: palette.lotusGray3,
		Normal: palette.lotusInk1,

		// Notification colors
		InfoFg:    palette.lotusTeal3,
		InfoBg:    palette.lotusBlue2,
		WarningFg: palette.lotusOrange2,
		WarningBg: palette.lotusYellow4,
		ErrorFg:   palette.lotusRed3,
		ErrorBg:   palette.lotusRed4,
	}
}
