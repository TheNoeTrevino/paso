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

// FilterBarMode - Filter bar chip navigation
// Ctrl+F (from NormalMode): Enter filter bar mode
// h / ← (left arrow)     : Move focus to previous chip
// l / → (right arrow)     : Move focus to next chip
// Enter                   : Open picker for focused chip (label/priority/type/assignee)
// x / Delete / Backspace  : Clear filter on focused chip (or Clear All if on last chip)
// Esc / q                 : Exit filter bar mode (return to NormalMode)
//
// Chips wrap around: moving left from the first chip goes to Clear All,
// moving right from Clear All goes to the first chip.
// Clearing a chip triggers an immediate task re-fetch.
