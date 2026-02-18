package renderers

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

var (
	datePickerHeaderStyle  lipgloss.Style
	datePickerWeekdayStyle lipgloss.Style
	datePickerDayStyle     lipgloss.Style
	datePickerCursorStyle  lipgloss.Style
	datePickerOnce         sync.Once
)

// InitDatePickerStyles initializes date picker styles with theme colors
// Thread-safe: uses sync.Once to ensure initialization happens exactly once
func InitDatePickerStyles() {
	datePickerOnce.Do(func() {
		datePickerHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(theme.Title))

		datePickerWeekdayStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Subtle))

		datePickerDayStyle = lipgloss.NewStyle().
			Padding(0, 1)

		datePickerCursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(theme.SelectedBg)).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)
	})
}

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
	for week := range 6 {
		for dayOfWeek := range 7 {
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

// lastActiveWeek returns the index of the last week (row) in the grid that contains
// at least one valid day number (> 0). Returns 0 if no days are found.
func lastActiveWeek(grid [][]int) int {
	for i := len(grid) - 1; i >= 0; i-- {
		for _, day := range grid[i] {
			if day > 0 {
				return i
			}
		}
	}
	return 0
}

// DatePickerContentHeight calculates the number of lines needed to render the date picker
// for the given month and year. This includes header, weekday labels, and week rows.
func DatePickerContentHeight(month time.Month, year int) int {
	// Calculate number of weeks needed
	grid := generateCalendarGrid(month, year)
	weekRows := lastActiveWeek(grid) + 1

	// Header (1 line) + blank line (1) + weekdays (1) + blank line (1) + week rows
	return 2 + 2 + weekRows
}

// RenderDatePicker renders the complete calendar UI with Lipgloss styling.
// It displays a header with month/year, weekday labels, and a calendar grid
// with cursor highlighting. The border and padding are applied by the layer wrapper.
// The calendar is self-sizing based on its content (7 columns x 6 chars per cell = 42 chars).
func RenderDatePicker(pickerState *state.DatePickerState, width int, height int) string {
	var content strings.Builder

	// Calendar content width: 7 cells * 6 chars per cell = 42 chars
	const calendarContentWidth = 42

	// Header: "FEBRUARY 2026" (uppercase, bold, centered)
	header := strings.ToUpper(fmt.Sprintf("%s %d", pickerState.CurrentMonth.String(), pickerState.CurrentYear))
	headerStyleWithWidth := datePickerHeaderStyle.Width(calendarContentWidth)
	content.WriteString(headerStyleWithWidth.Render(header) + "\n\n")

	// Weekday labels: each label is 6 chars wide to match cell width
	weekdays := "  Sun   Mon   Tue   Wed   Thu   Fri   Sat"
	content.WriteString(datePickerWeekdayStyle.Render(weekdays) + "\n\n")

	// Calendar grid - only render weeks that contain days
	grid := generateCalendarGrid(pickerState.CurrentMonth, pickerState.CurrentYear)
	lastWeek := lastActiveWeek(grid)

	for weekIdx, week := range grid {
		if weekIdx > lastWeek {
			break
		}
		var weekStr strings.Builder
		for _, day := range week {
			switch day {
			case 0:
				// Empty cell - 6 spaces to match styled cell width
				weekStr.WriteString("      ")
			case pickerState.CursorDay:
				// Cursor-highlighted day
				weekStr.WriteString(datePickerCursorStyle.Render(fmt.Sprintf(" %2d ", day)))
			default:
				// Normal day
				weekStr.WriteString(datePickerDayStyle.Render(fmt.Sprintf(" %2d ", day)))
			}
		}
		content.WriteString(weekStr.String() + "\n")
	}

	return content.String()
}
