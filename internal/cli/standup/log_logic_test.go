package standup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/config/colors"
)

func TestFormatLogQuiet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		logID    int
		expected string
	}{
		{
			name:     "positive ID",
			logID:    42,
			expected: "42\n",
		},
		{
			name:     "zero ID",
			logID:    0,
			expected: "0\n",
		},
		{
			name:     "large ID",
			logID:    999999,
			expected: "999999\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, FormatLogQuiet(tt.logID))
		})
	}
}

func TestFormatLogJSON(t *testing.T) {
	t.Parallel()

	result := &LogResult{
		ID:        1,
		ProjectID: 5,
		Content:   "Fixed the auth bug",
		CreatedAt: "2026-02-22 10:30:00",
	}

	output := FormatLogJSON(result)

	assert.Equal(t, true, output["success"])

	log, ok := output["log"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, 1, log["id"])
	assert.Equal(t, 5, log["project_id"])
	assert.Equal(t, "Fixed the auth bug", log["content"])
	assert.Equal(t, "2026-02-22 10:30:00", log["created_at"])
}

func TestFormatLogHuman(t *testing.T) {
	t.Parallel()

	testColorScheme := colors.ColorScheme{
		Normal: "#FFFFFF",
		Create: "#00FF00",
	}

	tests := []struct {
		name           string
		result         *LogResult
		expectContains []string
	}{
		{
			name: "standard log output",
			result: &LogResult{
				ID:        1,
				ProjectID: 5,
				Content:   "Fixed the auth bug",
				CreatedAt: "2026-02-22 10:30:00",
			},
			expectContains: []string{
				"Standup logged",
				"1",
				"Fixed the auth bug",
				"2026-02-22 10:30:00",
				"✓",
			},
		},
		{
			name: "long content gets truncated",
			result: &LogResult{
				ID:        2,
				ProjectID: 5,
				Content:   "This is a very long standup log message that should be truncated because it exceeds the maximum display length for human output",
				CreatedAt: "2026-02-22 14:00:00",
			},
			expectContains: []string{
				"Standup logged",
				"...",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatLogHuman(tt.result, testColorScheme)
			assert.NotEmpty(t, output)

			for _, expected := range tt.expectContains {
				assert.Contains(t, output, expected)
			}
		})
	}
}
