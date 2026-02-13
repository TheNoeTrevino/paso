package state

import "time"

// DatePickerState manages the date picker modal state.
// This modal allows users to select a date for a task.
type DatePickerState struct {
	// CurrentMonth is the currently displayed month (1=Jan, 12=Dec)
	CurrentMonth time.Month

	// CurrentYear is the currently displayed year
	CurrentYear int

	// SelectedDate is the user's final date selection
	SelectedDate time.Time

	// CursorDay is the currently highlighted day (1-31)
	CursorDay int

	// ReturnMode is the mode to return to after closing the picker
	ReturnMode Mode
}

// NewDatePickerState creates a new DatePickerState with default values.
// Defaults to today's date (time.Now()) for current month/year/day.
func NewDatePickerState() *DatePickerState {
	now := time.Now()
	return &DatePickerState{
		CurrentMonth: now.Month(),
		CurrentYear:  now.Year(),
		SelectedDate: now,
		CursorDay:    now.Day(),
		ReturnMode:   TicketFormMode,
	}
}

// Reset resets all state to default values (today's date).
func (s *DatePickerState) Reset() {
	now := time.Now()
	s.CurrentMonth = now.Month()
	s.CurrentYear = now.Year()
	s.SelectedDate = now
	s.CursorDay = now.Day()
	s.ReturnMode = TicketFormMode
}

// MoveDay moves the cursor by delta days (positive or negative).
// Automatically wraps to next/previous month when crossing boundaries.
// Handles edge cases:
//   - Day 31 → next month: clamps to valid day (e.g., Jan 31 + 1 day → Feb 1)
//   - Month boundaries: Dec 31 +1 → Jan 1 (increments year)
//   - Jan 1 -1 → Dec 31 (decrements year)
func (s *DatePickerState) MoveDay(delta int) {
	currentDate := time.Date(s.CurrentYear, s.CurrentMonth, s.CursorDay, 0, 0, 0, 0, time.Local)
	newDate := currentDate.AddDate(0, 0, delta)

	s.CurrentYear = newDate.Year()
	s.CurrentMonth = newDate.Month()
	s.CursorDay = newDate.Day()
}

// MoveWeek moves the cursor by delta weeks (7 * delta days).
// Uses the same wrapping logic as MoveDay.
// Edge case: First week up arrow → previous month's last week.
func (s *DatePickerState) MoveWeek(delta int) {
	s.MoveDay(7 * delta)
}

// NextMonth increments the month, wrapping Dec→Jan and incrementing year.
// Edge case: Jan 31 → Feb: clamps cursor to 28/29.
func (s *DatePickerState) NextMonth() {
	// Calculate next month and year
	newMonth := s.CurrentMonth + 1
	newYear := s.CurrentYear

	if newMonth > 12 {
		newMonth = 1
		newYear++
	}

	// Clamp cursor day to the last valid day of the new month
	lastDayOfNewMonth := time.Date(newYear, newMonth+1, 0, 0, 0, 0, 0, time.Local).Day()
	newDay := s.CursorDay
	if newDay > lastDayOfNewMonth {
		newDay = lastDayOfNewMonth
	}

	s.CurrentYear = newYear
	s.CurrentMonth = newMonth
	s.CursorDay = newDay
}

// PrevMonth decrements the month, wrapping Jan→Dec and decrementing year.
// Edge case: Same clamping logic as NextMonth (e.g., Mar 31 → Feb clamps to 28/29).
func (s *DatePickerState) PrevMonth() {
	// Calculate previous month and year
	newMonth := s.CurrentMonth - 1
	newYear := s.CurrentYear

	if newMonth < 1 {
		newMonth = 12
		newYear--
	}

	// Clamp cursor day to the last valid day of the new month
	lastDayOfNewMonth := time.Date(newYear, newMonth+1, 0, 0, 0, 0, 0, time.Local).Day()
	newDay := s.CursorDay
	if newDay > lastDayOfNewMonth {
		newDay = lastDayOfNewMonth
	}

	s.CurrentYear = newYear
	s.CurrentMonth = newMonth
	s.CursorDay = newDay
}
