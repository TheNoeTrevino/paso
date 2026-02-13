package state

import (
	"testing"
	"time"
)

// TestMoveDay_WrapToNextMonth tests moving forward across month boundaries
func TestMoveDay_WrapToNextMonth(t *testing.T) {
	tests := []struct {
		name          string
		startYear     int
		startMonth    time.Month
		startDay      int
		delta         int
		expectedYear  int
		expectedMonth time.Month
		expectedDay   int
	}{
		{
			name:          "Jan 31 + 1 day = Feb 1",
			startYear:     2024,
			startMonth:    time.January,
			startDay:      31,
			delta:         1,
			expectedYear:  2024,
			expectedMonth: time.February,
			expectedDay:   1,
		},
		{
			name:          "Jan 31 + 5 days = Feb 5",
			startYear:     2024,
			startMonth:    time.January,
			startDay:      31,
			delta:         5,
			expectedYear:  2024,
			expectedMonth: time.February,
			expectedDay:   5,
		},
		{
			name:          "Dec 31 + 1 day = Jan 1 (next year)",
			startYear:     2024,
			startMonth:    time.December,
			startDay:      31,
			delta:         1,
			expectedYear:  2025,
			expectedMonth: time.January,
			expectedDay:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &DatePickerState{
				CurrentYear:  tt.startYear,
				CurrentMonth: tt.startMonth,
				CursorDay:    tt.startDay,
			}

			state.MoveDay(tt.delta)

			if state.CurrentYear != tt.expectedYear {
				t.Errorf("expected year %d, got %d", tt.expectedYear, state.CurrentYear)
			}
			if state.CurrentMonth != tt.expectedMonth {
				t.Errorf("expected month %s, got %s", tt.expectedMonth, state.CurrentMonth)
			}
			if state.CursorDay != tt.expectedDay {
				t.Errorf("expected day %d, got %d", tt.expectedDay, state.CursorDay)
			}
		})
	}
}

// TestMoveDay_WrapToPreviousMonth tests moving backward across month boundaries
func TestMoveDay_WrapToPreviousMonth(t *testing.T) {
	tests := []struct {
		name          string
		startYear     int
		startMonth    time.Month
		startDay      int
		delta         int
		expectedYear  int
		expectedMonth time.Month
		expectedDay   int
	}{
		{
			name:          "Jan 1 - 1 day = Dec 31 (previous year)",
			startYear:     2024,
			startMonth:    time.January,
			startDay:      1,
			delta:         -1,
			expectedYear:  2023,
			expectedMonth: time.December,
			expectedDay:   31,
		},
		{
			name:          "Feb 1 - 1 day = Jan 31",
			startYear:     2024,
			startMonth:    time.February,
			startDay:      1,
			delta:         -1,
			expectedYear:  2024,
			expectedMonth: time.January,
			expectedDay:   31,
		},
		{
			name:          "Mar 1 - 5 days = Feb 25",
			startYear:     2024,
			startMonth:    time.March,
			startDay:      1,
			delta:         -5,
			expectedYear:  2024,
			expectedMonth: time.February,
			expectedDay:   25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &DatePickerState{
				CurrentYear:  tt.startYear,
				CurrentMonth: tt.startMonth,
				CursorDay:    tt.startDay,
			}

			state.MoveDay(tt.delta)

			if state.CurrentYear != tt.expectedYear {
				t.Errorf("expected year %d, got %d", tt.expectedYear, state.CurrentYear)
			}
			if state.CurrentMonth != tt.expectedMonth {
				t.Errorf("expected month %s, got %s", tt.expectedMonth, state.CurrentMonth)
			}
			if state.CursorDay != tt.expectedDay {
				t.Errorf("expected day %d, got %d", tt.expectedDay, state.CursorDay)
			}
		})
	}
}

// TestMoveWeek_WrapToNextMonth tests moving forward by weeks across month boundaries
func TestMoveWeek_WrapToNextMonth(t *testing.T) {
	tests := []struct {
		name          string
		startYear     int
		startMonth    time.Month
		startDay      int
		delta         int
		expectedYear  int
		expectedMonth time.Month
		expectedDay   int
	}{
		{
			name:          "Jan 25 + 1 week = Feb 1",
			startYear:     2024,
			startMonth:    time.January,
			startDay:      25,
			delta:         1,
			expectedYear:  2024,
			expectedMonth: time.February,
			expectedDay:   1,
		},
		{
			name:          "Dec 28 + 1 week = Jan 4 (next year)",
			startYear:     2024,
			startMonth:    time.December,
			startDay:      28,
			delta:         1,
			expectedYear:  2025,
			expectedMonth: time.January,
			expectedDay:   4,
		},
		{
			name:          "Jan 15 + 2 weeks = Jan 29",
			startYear:     2024,
			startMonth:    time.January,
			startDay:      15,
			delta:         2,
			expectedYear:  2024,
			expectedMonth: time.January,
			expectedDay:   29,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &DatePickerState{
				CurrentYear:  tt.startYear,
				CurrentMonth: tt.startMonth,
				CursorDay:    tt.startDay,
			}

			state.MoveWeek(tt.delta)

			if state.CurrentYear != tt.expectedYear {
				t.Errorf("expected year %d, got %d", tt.expectedYear, state.CurrentYear)
			}
			if state.CurrentMonth != tt.expectedMonth {
				t.Errorf("expected month %s, got %s", tt.expectedMonth, state.CurrentMonth)
			}
			if state.CursorDay != tt.expectedDay {
				t.Errorf("expected day %d, got %d", tt.expectedDay, state.CursorDay)
			}
		})
	}
}

// TestMoveWeek_WrapToPreviousMonth tests moving backward by weeks across month boundaries
func TestMoveWeek_WrapToPreviousMonth(t *testing.T) {
	tests := []struct {
		name          string
		startYear     int
		startMonth    time.Month
		startDay      int
		delta         int
		expectedYear  int
		expectedMonth time.Month
		expectedDay   int
	}{
		{
			name:          "Feb 5 - 1 week = Jan 29",
			startYear:     2024,
			startMonth:    time.February,
			startDay:      5,
			delta:         -1,
			expectedYear:  2024,
			expectedMonth: time.January,
			expectedDay:   29,
		},
		{
			name:          "Jan 3 - 1 week = Dec 27 (previous year)",
			startYear:     2024,
			startMonth:    time.January,
			startDay:      3,
			delta:         -1,
			expectedYear:  2023,
			expectedMonth: time.December,
			expectedDay:   27,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &DatePickerState{
				CurrentYear:  tt.startYear,
				CurrentMonth: tt.startMonth,
				CursorDay:    tt.startDay,
			}

			state.MoveWeek(tt.delta)

			if state.CurrentYear != tt.expectedYear {
				t.Errorf("expected year %d, got %d", tt.expectedYear, state.CurrentYear)
			}
			if state.CurrentMonth != tt.expectedMonth {
				t.Errorf("expected month %s, got %s", tt.expectedMonth, state.CurrentMonth)
			}
			if state.CursorDay != tt.expectedDay {
				t.Errorf("expected day %d, got %d", tt.expectedDay, state.CursorDay)
			}
		})
	}
}

// TestNextMonth_YearBoundary tests that NextMonth increments year when crossing Dec→Jan
func TestNextMonth_YearBoundary(t *testing.T) {
	state := &DatePickerState{
		CurrentYear:  2024,
		CurrentMonth: time.December,
		CursorDay:    15,
		cursorWant:   15,
	}

	state.NextMonth()

	if state.CurrentYear != 2025 {
		t.Errorf("expected year 2025, got %d", state.CurrentYear)
	}
	if state.CurrentMonth != time.January {
		t.Errorf("expected month January, got %s", state.CurrentMonth)
	}
	if state.CursorDay != 15 {
		t.Errorf("expected day 15, got %d", state.CursorDay)
	}
}

// TestPrevMonth_YearBoundary tests that PrevMonth decrements year when crossing Jan→Dec
func TestPrevMonth_YearBoundary(t *testing.T) {
	state := &DatePickerState{
		CurrentYear:  2024,
		CurrentMonth: time.January,
		CursorDay:    15,
		cursorWant:   15,
	}

	state.PrevMonth()

	if state.CurrentYear != 2023 {
		t.Errorf("expected year 2023, got %d", state.CurrentYear)
	}
	if state.CurrentMonth != time.December {
		t.Errorf("expected month December, got %s", state.CurrentMonth)
	}
	if state.CursorDay != 15 {
		t.Errorf("expected day 15, got %d", state.CursorDay)
	}
}

// TestNextMonth_CursorClamping tests that cursor clamps to valid day when moving to shorter month
func TestNextMonth_CursorClamping(t *testing.T) {
	tests := []struct {
		name          string
		startYear     int
		startMonth    time.Month
		startDay      int
		expectedYear  int
		expectedMonth time.Month
		expectedDay   int
	}{
		{
			name:          "Jan 31 → Feb clamps to Feb 29 (leap year)",
			startYear:     2024,
			startMonth:    time.January,
			startDay:      31,
			expectedYear:  2024,
			expectedMonth: time.February,
			expectedDay:   29,
		},
		{
			name:          "Jan 31 → Feb clamps to Feb 28 (non-leap year)",
			startYear:     2023,
			startMonth:    time.January,
			startDay:      31,
			expectedYear:  2023,
			expectedMonth: time.February,
			expectedDay:   28,
		},
		{
			name:          "Mar 31 → Apr clamps to Apr 30",
			startYear:     2024,
			startMonth:    time.March,
			startDay:      31,
			expectedYear:  2024,
			expectedMonth: time.April,
			expectedDay:   30,
		},
		{
			name:          "May 31 → Jun clamps to Jun 30",
			startYear:     2024,
			startMonth:    time.May,
			startDay:      31,
			expectedYear:  2024,
			expectedMonth: time.June,
			expectedDay:   30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &DatePickerState{
				CurrentYear:  tt.startYear,
				CurrentMonth: tt.startMonth,
				CursorDay:    tt.startDay,
				cursorWant:   tt.startDay,
			}

			state.NextMonth()

			if state.CurrentYear != tt.expectedYear {
				t.Errorf("expected year %d, got %d", tt.expectedYear, state.CurrentYear)
			}
			if state.CurrentMonth != tt.expectedMonth {
				t.Errorf("expected month %s, got %s", tt.expectedMonth, state.CurrentMonth)
			}
			if state.CursorDay != tt.expectedDay {
				t.Errorf("expected day %d, got %d", tt.expectedDay, state.CursorDay)
			}
		})
	}
}

// TestPrevMonth_CursorClamping tests that cursor clamps when moving backward to shorter month
func TestPrevMonth_CursorClamping(t *testing.T) {
	tests := []struct {
		name          string
		startYear     int
		startMonth    time.Month
		startDay      int
		expectedYear  int
		expectedMonth time.Month
		expectedDay   int
	}{
		{
			name:          "Mar 31 → Feb clamps to Feb 29 (leap year)",
			startYear:     2024,
			startMonth:    time.March,
			startDay:      31,
			expectedYear:  2024,
			expectedMonth: time.February,
			expectedDay:   29,
		},
		{
			name:          "Mar 31 → Feb clamps to Feb 28 (non-leap year)",
			startYear:     2023,
			startMonth:    time.March,
			startDay:      31,
			expectedYear:  2023,
			expectedMonth: time.February,
			expectedDay:   28,
		},
		{
			name:          "May 31 → Apr clamps to Apr 30",
			startYear:     2024,
			startMonth:    time.May,
			startDay:      31,
			expectedYear:  2024,
			expectedMonth: time.April,
			expectedDay:   30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &DatePickerState{
				CurrentYear:  tt.startYear,
				CurrentMonth: tt.startMonth,
				CursorDay:    tt.startDay,
				cursorWant:   tt.startDay,
			}

			state.PrevMonth()

			if state.CurrentYear != tt.expectedYear {
				t.Errorf("expected year %d, got %d", tt.expectedYear, state.CurrentYear)
			}
			if state.CurrentMonth != tt.expectedMonth {
				t.Errorf("expected month %s, got %s", tt.expectedMonth, state.CurrentMonth)
			}
			if state.CursorDay != tt.expectedDay {
				t.Errorf("expected day %d, got %d", tt.expectedDay, state.CursorDay)
			}
		})
	}
}

// TestLeapYear_February29 tests leap year handling for February 29
func TestLeapYear_February29(t *testing.T) {
	tests := []struct {
		name   string
		year   int
		isLeap bool
	}{
		{"2024 is leap year", 2024, true},
		{"2023 is not leap year", 2023, false},
		{"2000 is leap year", 2000, true},
		{"1900 is not leap year", 1900, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &DatePickerState{
				CurrentYear:  tt.year,
				CurrentMonth: time.February,
				CursorDay:    28,
			}

			state.MoveDay(1)

			if tt.isLeap {
				if state.CursorDay != 29 {
					t.Errorf("leap year %d: expected Feb 29, got %d", tt.year, state.CursorDay)
				}
				if state.CurrentMonth != time.February {
					t.Errorf("leap year %d: should still be in February", tt.year)
				}
			} else {
				if state.CursorDay != 1 {
					t.Errorf("non-leap year %d: expected Mar 1, got day %d", tt.year, state.CursorDay)
				}
				if state.CurrentMonth != time.March {
					t.Errorf("non-leap year %d: expected March, got %s", tt.year, state.CurrentMonth)
				}
			}
		})
	}
}

// TestCursorWant_MonthNavigation_RestoresDay tests that navigating to a shorter
// month and back restores the original day (sticky cursor behavior)
func TestCursorWant_MonthNavigation_RestoresDay(t *testing.T) {
	tests := []struct {
		name          string
		startYear     int
		startMonth    time.Month
		startDay      int
		expectedYear  int
		expectedMonth time.Month
		expectedDay   int
	}{
		{
			name:          "Jan 31 → Feb 28 → Jan 31 (non-leap year)",
			startYear:     2023,
			startMonth:    time.January,
			startDay:      31,
			expectedYear:  2023,
			expectedMonth: time.January,
			expectedDay:   31,
		},
		{
			name:          "Jan 31 → Feb 29 → Jan 31 (leap year)",
			startYear:     2024,
			startMonth:    time.January,
			startDay:      31,
			expectedYear:  2024,
			expectedMonth: time.January,
			expectedDay:   31,
		},
		{
			name:          "Mar 31 → Apr 30 → Mar 31",
			startYear:     2024,
			startMonth:    time.March,
			startDay:      31,
			expectedYear:  2024,
			expectedMonth: time.March,
			expectedDay:   31,
		},
		{
			name:          "May 31 → Jun 30 → May 31",
			startYear:     2024,
			startMonth:    time.May,
			startDay:      31,
			expectedYear:  2024,
			expectedMonth: time.May,
			expectedDay:   31,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &DatePickerState{
				CurrentYear:  tt.startYear,
				CurrentMonth: tt.startMonth,
				CursorDay:    tt.startDay,
				cursorWant:   tt.startDay,
			}

			state.NextMonth()
			state.PrevMonth()

			if state.CurrentYear != tt.expectedYear {
				t.Errorf("expected year %d, got %d", tt.expectedYear, state.CurrentYear)
			}
			if state.CurrentMonth != tt.expectedMonth {
				t.Errorf("expected month %s, got %s", tt.expectedMonth, state.CurrentMonth)
			}
			if state.CursorDay != tt.expectedDay {
				t.Errorf("expected day %d, got %d", tt.expectedDay, state.CursorDay)
			}
		})
	}
}

// TestCursorWant_MultipleMonthChanges tests that cursorWant persists across
// multiple month changes
func TestCursorWant_MultipleMonthChanges(t *testing.T) {
	state := &DatePickerState{
		CurrentYear:  2024,
		CurrentMonth: time.January,
		CursorDay:    31,
		cursorWant:   31,
	}

	state.NextMonth()
	if state.CursorDay != 29 {
		t.Errorf("Feb: expected day 29, got %d", state.CursorDay)
	}

	state.NextMonth()
	if state.CursorDay != 31 {
		t.Errorf("Mar: expected day 31 (restored), got %d", state.CursorDay)
	}

	state.NextMonth()
	if state.CursorDay != 30 {
		t.Errorf("Apr: expected day 30, got %d", state.CursorDay)
	}

	state.PrevMonth()
	if state.CursorDay != 31 {
		t.Errorf("Mar (back): expected day 31 (restored), got %d", state.CursorDay)
	}

	if state.cursorWant != 31 {
		t.Errorf("cursorWant should still be 31, got %d", state.cursorWant)
	}
}

// TestCursorWant_UpdatesOnDayMovement tests that MoveDay updates cursorWant
func TestCursorWant_UpdatesOnDayMovement(t *testing.T) {
	state := &DatePickerState{
		CurrentYear:  2024,
		CurrentMonth: time.January,
		CursorDay:    15,
		cursorWant:   15,
	}

	state.MoveDay(5)

	if state.cursorWant != 20 {
		t.Errorf("cursorWant should be 20 after MoveDay(5), got %d", state.cursorWant)
	}

	state.NextMonth()
	if state.CursorDay != 20 {
		t.Errorf("Feb: expected day 20 (from cursorWant), got %d", state.CursorDay)
	}
}

// TestCursorWant_UpdatesOnWeekMovement tests that MoveWeek updates cursorWant
func TestCursorWant_UpdatesOnWeekMovement(t *testing.T) {
	state := &DatePickerState{
		CurrentYear:  2024,
		CurrentMonth: time.January,
		CursorDay:    10,
		cursorWant:   10,
	}

	state.MoveWeek(2)

	if state.cursorWant != 24 {
		t.Errorf("cursorWant should be 24 after MoveWeek(2), got %d", state.cursorWant)
	}

	state.NextMonth()
	if state.CursorDay != 24 {
		t.Errorf("Feb: expected day 24 (from cursorWant), got %d", state.CursorDay)
	}
}

// TestCursorWant_DoesNotUpdateOnMonthChange tests that NextMonth/PrevMonth
// do NOT update cursorWant
func TestCursorWant_DoesNotUpdateOnMonthChange(t *testing.T) {
	state := &DatePickerState{
		CurrentYear:  2024,
		CurrentMonth: time.January,
		CursorDay:    31,
		cursorWant:   31,
	}

	state.NextMonth()

	if state.CursorDay != 29 {
		t.Errorf("CursorDay should be 29 (clamped), got %d", state.CursorDay)
	}
	if state.cursorWant != 31 {
		t.Errorf("cursorWant should still be 31, got %d", state.cursorWant)
	}

	state.NextMonth()

	if state.CursorDay != 31 {
		t.Errorf("CursorDay should be 31 (restored), got %d", state.CursorDay)
	}
	if state.cursorWant != 31 {
		t.Errorf("cursorWant should still be 31, got %d", state.cursorWant)
	}
}

// TestCursorWant_YearBoundary tests sticky cursor across year boundaries
func TestCursorWant_YearBoundary(t *testing.T) {
	state := &DatePickerState{
		CurrentYear:  2024,
		CurrentMonth: time.December,
		CursorDay:    31,
		cursorWant:   31,
	}

	state.NextMonth()

	if state.CurrentYear != 2025 {
		t.Errorf("expected year 2025, got %d", state.CurrentYear)
	}
	if state.CurrentMonth != time.January {
		t.Errorf("expected month January, got %s", state.CurrentMonth)
	}
	if state.CursorDay != 31 {
		t.Errorf("expected day 31 (preserved), got %d", state.CursorDay)
	}

	state.NextMonth()
	if state.CursorDay != 28 {
		t.Errorf("expected day 28 (clamped), got %d", state.CursorDay)
	}

	state.PrevMonth()
	if state.CursorDay != 31 {
		t.Errorf("expected day 31 (restored), got %d", state.CursorDay)
	}

	state.PrevMonth()
	if state.CurrentYear != 2024 {
		t.Errorf("expected year 2024, got %d", state.CurrentYear)
	}
	if state.CurrentMonth != time.December {
		t.Errorf("expected month December, got %s", state.CurrentMonth)
	}
	if state.CursorDay != 31 {
		t.Errorf("expected day 31 (preserved), got %d", state.CursorDay)
	}
}

// TestCursorWant_UserChangesMind tests that user can override cursorWant
// by navigating to a different day
func TestCursorWant_UserChangesMind(t *testing.T) {
	state := &DatePickerState{
		CurrentYear:  2024,
		CurrentMonth: time.January,
		CursorDay:    31,
		cursorWant:   31,
	}

	state.NextMonth()
	if state.CursorDay != 29 {
		t.Errorf("Feb: expected day 29, got %d", state.CursorDay)
	}

	state.MoveDay(-1)
	if state.CursorDay != 28 {
		t.Errorf("expected day 28 after MoveDay(-1), got %d", state.CursorDay)
	}

	if state.cursorWant != 28 {
		t.Errorf("cursorWant should be 28 (updated), got %d", state.cursorWant)
	}

	state.PrevMonth()
	if state.CursorDay != 28 {
		t.Errorf("Jan: expected day 28 (new cursorWant), got %d", state.CursorDay)
	}
}
