package components

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

func TestFormatRelativeDueDate_NilDate(t *testing.T) {
	t.Parallel()

	text, color := FormatRelativeDueDate(nil)

	assert.Empty(t, text)
	assert.Empty(t, color)
}

func TestFormatRelativeDueDate_AllPaths(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		dueDate   time.Time
		wantText  string
		wantColor string
	}{
		{
			name:      "overdue by 1 day",
			dueDate:   time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC),
			wantText:  "Overdue by 1 day",
			wantColor: theme.ErrorFg,
		},
		{
			name:      "overdue by 5 days",
			dueDate:   time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
			wantText:  "Overdue by 5 days",
			wantColor: theme.ErrorFg,
		},
		{
			name:      "overdue by 30 days",
			dueDate:   time.Date(2026, 2, 13, 0, 0, 0, 0, time.UTC),
			wantText:  "Overdue by 30 days",
			wantColor: theme.ErrorFg,
		},
		{
			name:      "due today",
			dueDate:   time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			wantText:  "Due today",
			wantColor: theme.WarningFg,
		},
		{
			name:      "due tomorrow",
			dueDate:   time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
			wantText:  "Due tomorrow",
			wantColor: theme.WarningFg,
		},
		{
			name:      "due in 2 days",
			dueDate:   time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC),
			wantText:  "Due in 2 days",
			wantColor: theme.WarningFg,
		},
		{
			name:      "due in 3 days",
			dueDate:   time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
			wantText:  "Due in 3 days",
			wantColor: theme.WarningFg,
		},
		{
			name:      "due in 4 days - switches to highlight",
			dueDate:   time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
			wantText:  "Due in 4 days",
			wantColor: theme.Highlight,
		},
		{
			name:      "due in 10 days",
			dueDate:   time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC),
			wantText:  "Due in 10 days",
			wantColor: theme.Highlight,
		},
		{
			name:      "due in 365 days",
			dueDate:   time.Date(2027, 3, 15, 0, 0, 0, 0, time.UTC),
			wantText:  "Due in 365 days",
			wantColor: theme.Highlight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dueDate := tt.dueDate
			text, color := formatRelativeDueDateFrom(&dueDate, now)

			assert.Equal(t, tt.wantText, text)
			assert.Equal(t, tt.wantColor, color)
		})
	}
}

func TestFormatRelativeDueDate_TimeOfDayIgnored(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 15, 23, 59, 59, 0, time.UTC)
	dueDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	text, _ := formatRelativeDueDateFrom(&dueDate, now)

	assert.Equal(t, "Due today", text, "time of day should not affect day comparison")
}

func TestFormatRelativeDueDate_CrossMonthBoundary(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	text, color := formatRelativeDueDateFrom(&dueDate, now)

	assert.Equal(t, "Due tomorrow", text)
	assert.Equal(t, theme.WarningFg, color)
}

func TestFormatRelativeDueDate_CrossYearBoundary(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

	text, color := formatRelativeDueDateFrom(&dueDate, now)

	assert.Equal(t, "Due tomorrow", text)
	assert.Equal(t, theme.WarningFg, color)
}
