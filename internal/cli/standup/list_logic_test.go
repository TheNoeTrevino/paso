package standup

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestFormatListQuiet(t *testing.T) {
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
			name: "single log",
			logs: []models.StandupLog{
				{ID: 1},
			},
			expected: []string{"1"},
		},
		{
			name: "multiple logs",
			logs: []models.StandupLog{
				{ID: 3},
				{ID: 1},
				{ID: 2},
			},
			expected: []string{"3", "1", "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, FormatListQuiet(tt.logs))
		})
	}
}

func TestFormatListJSON(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 22, 10, 30, 0, 0, time.UTC)

	logs := []models.StandupLog{
		{
			ID:        1,
			ProjectID: 5,
			Content:   "Fixed auth bug",
			CreatedAt: now,
		},
		{
			ID:        2,
			ProjectID: 5,
			Content:   "Added tests",
			CreatedAt: now.Add(time.Hour),
		},
	}

	output := FormatListJSON(logs)

	assert.Equal(t, true, output["success"])

	items, ok := output["logs"].([]map[string]any)
	assert.True(t, ok)
	assert.Len(t, items, 2)
	assert.Equal(t, 1, items[0]["id"])
	assert.Equal(t, "Fixed auth bug", items[0]["content"])
	assert.Equal(t, 2, items[1]["id"])
	assert.Equal(t, "Added tests", items[1]["content"])
}

func TestFormatListJSON_Empty(t *testing.T) {
	t.Parallel()

	output := FormatListJSON([]models.StandupLog{})

	assert.Equal(t, true, output["success"])
	items, ok := output["logs"].([]map[string]any)
	assert.True(t, ok)
	assert.Empty(t, items)
}

func TestFormatListHuman(t *testing.T) {
	t.Parallel()

	testColorScheme := colors.ColorScheme{
		Normal: "#FFFFFF",
		Create: "#00FF00",
		Subtle: "#888888",
	}

	now := time.Date(2026, 2, 22, 10, 30, 0, 0, time.UTC)

	logs := []models.StandupLog{
		{
			ID:        1,
			ProjectID: 5,
			Content:   "Fixed auth bug",
			CreatedAt: now,
		},
		{
			ID:        2,
			ProjectID: 5,
			Content:   "Added unit tests\nfor the login flow",
			CreatedAt: now.Add(2 * time.Hour),
		},
	}

	output := FormatListHuman(logs, testColorScheme)

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "#1")
	assert.Contains(t, output, "#2")
	assert.Contains(t, output, "Fixed auth bug")
	assert.Contains(t, output, "Added unit tests")
	assert.Contains(t, output, "for the login flow")
	assert.Contains(t, output, "2 standup log(s)")
}
