package standup

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestComputeSince(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 22, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		days     int
		weeks    int
		expected time.Time
	}{
		{
			name:     "1 day",
			days:     1,
			weeks:    0,
			expected: time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1 week",
			days:     0,
			weeks:    1,
			expected: time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "3 days and 1 week (10 days total)",
			days:     3,
			weeks:    1,
			expected: time.Date(2026, 2, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2 weeks",
			days:     0,
			weeks:    2,
			expected: time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "0 days 0 weeks returns start of today",
			days:     0,
			weeks:    0,
			expected: time.Date(2026, 2, 22, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "crosses month boundary",
			days:     25,
			weeks:    0,
			expected: time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ComputeSince(now, tt.days, tt.weeks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComputeSince_StartsAtMidnight(t *testing.T) {
	t.Parallel()

	// Even if "now" is mid-day, since should always be midnight
	now := time.Date(2026, 2, 22, 18, 45, 30, 0, time.UTC)
	result := ComputeSince(now, 1, 0)

	assert.Equal(t, 0, result.Hour())
	assert.Equal(t, 0, result.Minute())
	assert.Equal(t, 0, result.Second())
}

func TestGroupLogsByDate(t *testing.T) {
	t.Parallel()

	day1Morning := time.Date(2026, 2, 22, 9, 0, 0, 0, time.UTC)
	day1Afternoon := time.Date(2026, 2, 22, 14, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 2, 21, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		logs           []models.StandupLog
		expectedGroups int
		expectedDates  []string
		expectedCounts []int
	}{
		{
			name:           "empty logs",
			logs:           []models.StandupLog{},
			expectedGroups: 0,
		},
		{
			name: "single day",
			logs: []models.StandupLog{
				{ID: 2, CreatedAt: day1Afternoon},
				{ID: 1, CreatedAt: day1Morning},
			},
			expectedGroups: 1,
			expectedDates:  []string{"Sun, Feb 22 2026"},
			expectedCounts: []int{2},
		},
		{
			name: "multiple days",
			logs: []models.StandupLog{
				{ID: 3, CreatedAt: day1Afternoon},
				{ID: 2, CreatedAt: day1Morning},
				{ID: 1, CreatedAt: day2},
			},
			expectedGroups: 2,
			expectedDates:  []string{"Sun, Feb 22 2026", "Sat, Feb 21 2026"},
			expectedCounts: []int{2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			groups := GroupLogsByDate(tt.logs)
			assert.Len(t, groups, tt.expectedGroups)

			for i, g := range groups {
				if i < len(tt.expectedDates) {
					assert.Equal(t, tt.expectedDates[i], g.Date)
				}
				if i < len(tt.expectedCounts) {
					assert.Len(t, g.Logs, tt.expectedCounts[i])
				}
			}
		})
	}
}

func TestGroupLogsByDate_PreservesLogOrder(t *testing.T) {
	t.Parallel()

	// Logs come in newest-first from the DB query
	logs := []models.StandupLog{
		{ID: 3, CreatedAt: time.Date(2026, 2, 22, 15, 0, 0, 0, time.UTC)},
		{ID: 2, CreatedAt: time.Date(2026, 2, 22, 10, 0, 0, 0, time.UTC)},
		{ID: 1, CreatedAt: time.Date(2026, 2, 22, 8, 0, 0, 0, time.UTC)},
	}

	groups := GroupLogsByDate(logs)
	assert.Len(t, groups, 1)
	assert.Equal(t, 3, groups[0].Logs[0].ID)
	assert.Equal(t, 2, groups[0].Logs[1].ID)
	assert.Equal(t, 1, groups[0].Logs[2].ID)
}

func TestFormatGenerateQuiet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		logs     []models.StandupLog
		expected []string
	}{
		{
			name:     "empty logs",
			logs:     []models.StandupLog{},
			expected: []string{},
		},
		{
			name: "multiple logs",
			logs: []models.StandupLog{
				{ID: 3},
				{ID: 1},
			},
			expected: []string{"3", "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, FormatGenerateQuiet(tt.logs))
		})
	}
}

func TestFormatGenerateJSON(t *testing.T) {
	t.Parallel()

	since := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 2, 22, 14, 0, 0, 0, time.UTC)

	logs := []models.StandupLog{
		{
			ID:        2,
			ProjectID: 5,
			Content:   "Added tests",
			CreatedAt: time.Date(2026, 2, 22, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:        1,
			ProjectID: 5,
			Content:   "Fixed bug",
			CreatedAt: time.Date(2026, 2, 21, 9, 0, 0, 0, time.UTC),
		},
	}

	output := FormatGenerateJSON(logs, since, until)

	assert.Equal(t, true, output["success"])
	assert.Equal(t, 2, output["count"])
	assert.Equal(t, "2026-02-15T00:00:00Z", output["since"])
	assert.Equal(t, "2026-02-22T14:00:00Z", output["until"])

	groups, ok := output["groups"].([]map[string]any)
	assert.True(t, ok)
	assert.Len(t, groups, 2)
	assert.Equal(t, "Sun, Feb 22 2026", groups[0]["date"])
	assert.Equal(t, "Sat, Feb 21 2026", groups[1]["date"])
}

func TestFormatGenerateHuman(t *testing.T) {
	t.Parallel()

	testColorScheme := colors.ColorScheme{
		Normal: "#FFFFFF",
		Title:  "#FF00FF",
		Subtle: "#888888",
	}

	since := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 2, 22, 14, 0, 0, 0, time.UTC)

	logs := []models.StandupLog{
		{
			ID:        2,
			ProjectID: 5,
			Content:   "Added unit tests",
			CreatedAt: time.Date(2026, 2, 22, 14, 0, 0, 0, time.UTC),
		},
		{
			ID:        1,
			ProjectID: 5,
			Content:   "Fixed auth bug\nin login flow",
			CreatedAt: time.Date(2026, 2, 21, 9, 0, 0, 0, time.UTC),
		},
	}

	output := FormatGenerateHuman(logs, since, until, testColorScheme)

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Standup report")
	assert.Contains(t, output, "2 log(s)")
	assert.Contains(t, output, "Feb 22 2026")
	assert.Contains(t, output, "Feb 21 2026")
	assert.Contains(t, output, "#2")
	assert.Contains(t, output, "#1")
	assert.Contains(t, output, "Added unit tests")
	assert.Contains(t, output, "Fixed auth bug")
	assert.Contains(t, output, "in login flow")
}

func TestFormatGenerateHuman_MultilineContent(t *testing.T) {
	t.Parallel()

	testColorScheme := colors.ColorScheme{
		Normal: "#FFFFFF",
		Title:  "#FF00FF",
		Subtle: "#888888",
	}

	since := time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 2, 22, 14, 0, 0, 0, time.UTC)

	logs := []models.StandupLog{
		{
			ID:        1,
			ProjectID: 5,
			Content:   "Line one\nLine two\nLine three",
			CreatedAt: time.Date(2026, 2, 22, 10, 0, 0, 0, time.UTC),
		},
	}

	output := FormatGenerateHuman(logs, since, until, testColorScheme)

	assert.Contains(t, output, "Line one")
	assert.Contains(t, output, "Line two")
	assert.Contains(t, output, "Line three")
}
