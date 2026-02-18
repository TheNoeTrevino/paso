package state

import "time"

// DatePickerState manages the date picker modal state.
// This modal allows users to select a date for a task.
//
// Navigation uses a "cursor want" pattern (similar to Vim): when moving between
// months, the picker remembers which day number the user wanted and tries to
// return to it when possible. For example, if the user is on Jan 31 and navigates
// to February (clamping to Feb 28), then navigates back to January, they will
// return to Jan 31.
type DatePickerState struct {
	// CurrentMonth is the currently displayed month (1=Jan, 12=Dec)
	CurrentMonth time.Month

	// CurrentYear is the currently displayed year
	CurrentYear int

	// SelectedDate is the user's final date selection
	SelectedDate time.Time

	// CursorDay is the currently highlighted day (1-31)
	CursorDay int

	// cursorWant is the desired day number the user wants to be on.
	// Updated when user navigates with h/j/k/l (day/week movement).
	// Used by NextMonth/PrevMonth to restore position when possible.
	cursorWant int

	// ReturnMode is the mode to return to after closing the picker
	ReturnMode Mode
}

// NewDatePickerState creates a new DatePickerState with default values.
// Defaults to today's date (time.Now()) for current month/year/day.
// Initializes cursorWant to today's day for sticky navigation.
func NewDatePickerState() *DatePickerState {
	now := time.Now()
	return &DatePickerState{
		CurrentMonth: now.Month(),
		CurrentYear:  now.Year(),
		SelectedDate: now,
		CursorDay:    now.Day(),
		cursorWant:   now.Day(),
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
	s.cursorWant = now.Day()
	s.ReturnMode = TicketFormMode
}

// MoveDay moves the cursor by delta days (positive or negative).
// Automatically wraps to next/previous month when crossing boundaries.
// Updates cursorWant to the new day position (sticky navigation).
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
	s.cursorWant = newDate.Day()
}

// MoveWeek moves the cursor by delta weeks (7 * delta days).
// Uses the same wrapping logic as MoveDay.
// Updates cursorWant via MoveDay (sticky navigation).
// Edge case: First week up arrow → previous month's last week.
func (s *DatePickerState) MoveWeek(delta int) {
	s.MoveDay(7 * delta)
}

// navigateMonth moves the month by delta months (positive=forward, negative=backward).
// Handles year wrapping automatically via time.Date normalization.
// Clamps CursorDay to the last valid day of the new month using cursorWant.
// Does NOT update cursorWant (preserves the desired day position for sticky navigation).
func (s *DatePickerState) navigateMonth(delta int) {
	// time.Date automatically normalizes out-of-range months
	// (e.g., month 13 → Jan next year, month 0 → Dec prev year)
	newDate := time.Date(s.CurrentYear, s.CurrentMonth+time.Month(delta), 1, 0, 0, 0, 0, time.Local)
	newYear := newDate.Year()
	newMonth := newDate.Month()

	// Clamp cursorWant to the last valid day of the new month
	lastDayOfNewMonth := time.Date(newYear, newMonth+1, 0, 0, 0, 0, 0, time.Local).Day()
	newDay := min(s.cursorWant, lastDayOfNewMonth)

	s.CurrentYear = newYear
	s.CurrentMonth = newMonth
	s.CursorDay = newDay
}

// NextMonth increments the month, wrapping Dec→Jan and incrementing year.
// Uses cursorWant to maintain sticky navigation: if the user was on day 31
// and moved to a shorter month (clamping to day 28), then back to a 31-day
// month, they will return to day 31.
// Does NOT update cursorWant (preserves the desired day position).
func (s *DatePickerState) NextMonth() {
	s.navigateMonth(1)
}

// PrevMonth decrements the month, wrapping Jan→Dec and decrementing year.
// Uses cursorWant to maintain sticky navigation: if the user was on day 31
// and moved to a shorter month (clamping to day 28), then back to a 31-day
// month, they will return to day 31.
// Does NOT update cursorWant (preserves the desired day position).
func (s *DatePickerState) PrevMonth() {
	s.navigateMonth(-1)
}

// InitFromDate initializes the picker to display and select the given date.
// If date is nil, defaults to today's date.
// This is useful when editing a task with an existing due date.
func (s *DatePickerState) InitFromDate(date *time.Time) {
	var d time.Time
	if date == nil {
		d = time.Now()
	} else {
		d = *date
	}

	s.CurrentMonth = d.Month()
	s.CurrentYear = d.Year()
	s.CursorDay = d.Day()
	s.cursorWant = d.Day()
	s.SelectedDate = d
}
