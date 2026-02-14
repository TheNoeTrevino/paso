package tui

// This file documents keyboard shortcuts for different TUI modes.
// Each mode section lists the key bindings available in that interaction mode.

// DatePickerMode - Date picker for selecting dates
// Arrow keys (↑↓←→) or hjkl: Navigate calendar (up/down moves by week)
// [ / ]                     : Previous/Next month
// Enter                     : Select highlighted date
// Esc                       : Cancel (return zero time)
//
// Navigation uses "cursor want" pattern (like Vim): if you're on day 31 and
// move to a shorter month (Feb), the picker remembers day 31 and returns to it
// when you navigate back to a month with 31 days. The desired day is updated
// only when you use arrow keys/hjkl, not when changing months with [ / ].
