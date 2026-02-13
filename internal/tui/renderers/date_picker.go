package renderers

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(theme.Highlight))

	weekdayStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Subtle))

	dayStyle = lipgloss.NewStyle().
			Padding(0, 1)

	cursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(theme.Highlight)).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.Highlight))
)

// generateCalendarGrid generates a 6x7 grid representing a calendar month.
// Each cell contains the day number (1-31), with 0 representing empty cells
// before the month starts. The grid always has 6 rows (weeks) and 7 columns (days).
func generateCalendarGrid(month time.Month, year int) [][]int {
	grid := make([][]int, 6)
	for i := range grid {
		grid[i] = make([]int, 7)
	}

	// Get the weekday of the first day of the month (0=Sunday, 6=Saturday)
	weekdayOffset := int(monthStartWeekday(month, year))

	// Get the number of days in the month
	daysInMonthCount := daysInMonth(month, year)

	// Fill the grid with day numbers
	day := 1
	for week := 0; week < 6; week++ {
		for dayOfWeek := 0; dayOfWeek < 7; dayOfWeek++ {
			// Skip empty cells before the month starts
			if week == 0 && dayOfWeek < weekdayOffset {
				grid[week][dayOfWeek] = 0
				continue
			}

			// Stop when we've filled all days in the month
			if day > daysInMonthCount {
				grid[week][dayOfWeek] = 0
				continue
			}

			grid[week][dayOfWeek] = day
			day++
		}
	}

	return grid
}

// daysInMonth returns the number of days in the given month and year.
// It correctly handles leap years for February.
func daysInMonth(month time.Month, year int) int {
	// Get the 0th day of the next month, which gives us the last day of the current month
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}

// monthStartWeekday returns the weekday of the first day of the given month and year.
// Returns 0 for Sunday, 1 for Monday, ..., 6 for Saturday.
func monthStartWeekday(month time.Month, year int) int {
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	return int(firstDay.Weekday())
}

// RenderDatePicker renders the complete calendar UI with Lipgloss styling.
// It displays a header with month/year, weekday labels, and a calendar grid
// with cursor highlighting and a rounded border.
func RenderDatePicker(pickerState *state.DatePickerState, width int, height int) string {
	var content strings.Builder

	// Header: "FEBRUARY 2026" (uppercase, bold, centered)
	header := strings.ToUpper(fmt.Sprintf("%s %d", pickerState.CurrentMonth.String(), pickerState.CurrentYear))
	headerWidth := width - 4 // Account for border
	headerStyleWithWidth := headerStyle.Width(headerWidth)
	content.WriteString(headerStyleWithWidth.Render(header) + "\n\n")

	// Weekday labels: "Sun  Mon  Tue  Wed  Thu  Fri  Sat"
	weekdays := " Sun  Mon  Tue  Wed  Thu  Fri  Sat"
	content.WriteString(weekdayStyle.Render(weekdays) + "\n\n")

	// Calendar grid
	grid := generateCalendarGrid(pickerState.CurrentMonth, pickerState.CurrentYear)
	for _, week := range grid {
		var weekStr strings.Builder
		for _, day := range week {
			if day == 0 {
				// Empty cell - 5 spaces to maintain alignment
				weekStr.WriteString("     ")
			} else if day == pickerState.CursorDay {
				// Cursor-highlighted day
				weekStr.WriteString(cursorStyle.Render(fmt.Sprintf(" %2d ", day)))
			} else {
				// Normal day
				weekStr.WriteString(dayStyle.Render(fmt.Sprintf(" %2d ", day)))
			}
		}
		content.WriteString(weekStr.String() + "\n")
	}

	// Wrap in rounded border
	return borderStyle.Render(content.String())
}
