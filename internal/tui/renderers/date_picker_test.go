package renderers

import (
	"testing"
	"time"
)

func TestDaysInMonth(t *testing.T) {
	tests := []struct {
		name     string
		month    time.Month
		year     int
		expected int
	}{
		{"January", time.January, 2026, 31},
		{"February (non-leap)", time.February, 2026, 28},
		{"February (leap year)", time.February, 2024, 29},
		{"March", time.March, 2026, 31},
		{"April", time.April, 2026, 30},
		{"May", time.May, 2026, 31},
		{"June", time.June, 2026, 30},
		{"July", time.July, 2026, 31},
		{"August", time.August, 2026, 31},
		{"September", time.September, 2026, 30},
		{"October", time.October, 2026, 31},
		{"November", time.November, 2026, 30},
		{"December", time.December, 2026, 31},
		{"February (century non-leap)", time.February, 1900, 28},
		{"February (century leap)", time.February, 2000, 29},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := daysInMonth(tt.month, tt.year)
			if result != tt.expected {
				t.Errorf("daysInMonth(%v, %d) = %d; want %d", tt.month, tt.year, result, tt.expected)
			}
		})
	}
}

func TestMonthStartWeekday(t *testing.T) {
	tests := []struct {
		name     string
		month    time.Month
		year     int
		expected int
	}{
		{"January 2026 (Wednesday)", time.January, 2026, 4},
		{"February 2026 (Sunday)", time.February, 2026, 0},
		{"February 2024 (Thursday)", time.February, 2024, 4},
		{"March 2024 (Friday)", time.March, 2024, 5},
		{"December 2025 (Monday)", time.December, 2025, 1},
		{"July 2024 (Monday)", time.July, 2024, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := monthStartWeekday(tt.month, tt.year)
			if result != tt.expected {
				t.Errorf("monthStartWeekday(%v, %d) = %d; want %d", tt.month, tt.year, result, tt.expected)
			}
		})
	}
}

func TestGenerateCalendarGrid(t *testing.T) {
	tests := []struct {
		name     string
		month    time.Month
		year     int
		validate func(*testing.T, [][]int)
	}{
		{
			name:  "January 2026",
			month: time.January,
			year:  2026,
			validate: func(t *testing.T, grid [][]int) {
				// January 2026 starts on Thursday (weekday 4)
				// Grid should have 6 rows and 7 columns
				if len(grid) != 6 {
					t.Errorf("Expected 6 rows, got %d", len(grid))
				}
				for i := range grid {
					if len(grid[i]) != 7 {
						t.Errorf("Row %d: expected 7 columns, got %d", i, len(grid[i]))
					}
				}

				// First 4 cells (Sun-Wed) should be 0
				for i := range 4 {
					if grid[0][i] != 0 {
						t.Errorf("grid[0][%d] should be 0, got %d", i, grid[0][i])
					}
				}

				// Thursday (index 4) should be day 1
				if grid[0][4] != 1 {
					t.Errorf("grid[0][4] should be 1, got %d", grid[0][4])
				}

				// Friday (index 5) should be day 2
				if grid[0][5] != 2 {
					t.Errorf("grid[0][5] should be 2, got %d", grid[0][5])
				}

				// Saturday (index 6) should be day 3
				if grid[0][6] != 3 {
					t.Errorf("grid[0][6] should be 3, got %d", grid[0][6])
				}

				// Check that we have exactly 31 days
				dayCount := 0
				for week := range 6 {
					for day := range 7 {
						if grid[week][day] > 0 {
							dayCount++
						}
					}
				}
				if dayCount != 31 {
					t.Errorf("Expected 31 days, got %d", dayCount)
				}

				// Last day (31) should be in the correct position
				// January 31, 2026 is a Saturday
				if grid[4][6] != 31 {
					t.Errorf("grid[4][6] should be 31, got %d", grid[4][6])
				}
			},
		},
		{
			name:  "February 2024 (leap year)",
			month: time.February,
			year:  2024,
			validate: func(t *testing.T, grid [][]int) {
				// February 2024 starts on Thursday (weekday 4)
				// Should have 29 days (leap year)

				// First 4 cells should be 0
				for i := range 4 {
					if grid[0][i] != 0 {
						t.Errorf("grid[0][%d] should be 0, got %d", i, grid[0][i])
					}
				}

				// Thursday should be day 1
				if grid[0][4] != 1 {
					t.Errorf("grid[0][4] should be 1, got %d", grid[0][4])
				}

				// Check that we have exactly 29 days
				dayCount := 0
				for week := range 6 {
					for day := range 7 {
						if grid[week][day] > 0 {
							dayCount++
						}
					}
				}
				if dayCount != 29 {
					t.Errorf("Expected 29 days (leap year), got %d", dayCount)
				}

				// Last day (29) should be in the correct position
				// February 29, 2024 is a Thursday
				if grid[4][4] != 29 {
					t.Errorf("grid[4][4] should be 29, got %d", grid[4][4])
				}
			},
		},
		{
			name:  "February 2026 (non-leap year, starts Sunday)",
			month: time.February,
			year:  2026,
			validate: func(t *testing.T, grid [][]int) {
				// February 2026 starts on Sunday (weekday 0)
				// Should have 28 days (non-leap year)

				// Sunday should be day 1
				if grid[0][0] != 1 {
					t.Errorf("grid[0][0] should be 1, got %d", grid[0][0])
				}

				// Monday should be day 2
				if grid[0][1] != 2 {
					t.Errorf("grid[0][1] should be 2, got %d", grid[0][1])
				}

				// Check that we have exactly 28 days
				dayCount := 0
				for week := range 6 {
					for day := range 7 {
						if grid[week][day] > 0 {
							dayCount++
						}
					}
				}
				if dayCount != 28 {
					t.Errorf("Expected 28 days (non-leap year), got %d", dayCount)
				}

				// Last day (28) should be in the correct position
				// February 28, 2026 is a Saturday
				if grid[3][6] != 28 {
					t.Errorf("grid[3][6] should be 28, got %d", grid[3][6])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := generateCalendarGrid(tt.month, tt.year)
			tt.validate(t, grid)
		})
	}
}
