package colors

// Dragon returns the Kanagawa Dragon color scheme (dark theme with warm earth tones)
func Dragon() *ColorScheme {
	return &ColorScheme{
		Preset: "dragon",

		// Primary accent color
		Accent: palette.dragonViolet,

		Background: palette.dragonBlack1,

		// Semantic colors
		Create:  palette.dragonGreen2,
		Edit:    palette.dragonBlue2,
		Delete:  palette.dragonRed,
		Blocked: palette.dragonRed, // Use dragonRed for blocked indicator

		// UI element colors
		ColumnBorder:     palette.dragonBlack6,
		ColumnBackground: palette.roninYellow,
		TaskBorder:       palette.dragonBlack4,
		TaskBackground:   palette.dragonBlack3,
		SelectedBorder:   palette.dragonAqua,
		SelectedBg:       palette.waveBlue1,

		// Text colors
		Title:  palette.dragonBlue2,
		Subtle: palette.dragonAsh,
		Normal: palette.dragonWhite,

		// Notification colors
		InfoFg:    palette.dragonBlue,
		InfoBg:    palette.winterBlue,
		WarningFg: palette.roninYellow,
		WarningBg: palette.winterYellow,
		ErrorFg:   palette.samuraiRed,
		ErrorBg:   palette.winterRed,

		// Status bar
		StatusBarBg:   palette.dragonViolet, // Matches accent
		StatusBarText: palette.dragonWhite,  // Matches normal text
	}
}
