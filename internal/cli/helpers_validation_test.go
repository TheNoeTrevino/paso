package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateColorHex_Valid(t *testing.T) {
	t.Parallel()
	tests := []string{
		"#FF0000", // Red
		"#00FF00", // Green
		"#0000FF", // Blue
		"#FFFFFF", // White
		"#000000", // Black
		"#FF5733", // Random color
		"#ff5733", // Lowercase (should work)
		"#AbCdEf", // Mixed case
	}

	for _, color := range tests {
		t.Run(color, func(t *testing.T) {
			err := ValidateColorHex(color)
			assert.NoError(t, err, "Color should be valid: %s", color)
		})
	}
}

func TestValidateColorHex_Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		color       string
		description string
	}{
		{"FF0000", "missing # prefix"},
		{"#FFF", "too short (3 chars)"},
		{"#FF00000", "too long (7 chars)"},
		{"#GGGGGG", "invalid hex characters"},
		{"#FF00G0", "one invalid character"},
		{"#FF 000", "contains space"},
		{"", "empty string"},
		{"#", "only # symbol"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			err := ValidateColorHex(tt.color)
			assert.Error(t, err)
		})
	}
}

func TestParsePriority_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected int
	}{
		{"trivial", 1},
		{"low", 2},
		{"medium", 3},
		{"high", 4},
		{"critical", 5},
		// Test case insensitivity
		{"TRIVIAL", 1},
		{"Low", 2},
		{"MeDiUm", 3},
		{"HIGH", 4},
		{"Critical", 5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParsePriority(tt.input)
			assert.NoError(t, err, "ParsePriority should not return error for: %s", tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParsePriority_Invalid(t *testing.T) {
	t.Parallel()
	tests := []string{
		"invalid",
		"normal",
		"urgent",
		"",
		"123",
		"trivial ",
		" low",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParsePriority(input)
			assert.Error(t, err)
		})
	}
}

func TestParseTaskType_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected int
	}{
		{"task", 1},
		{"feature", 2},
		// Test case insensitivity
		{"TASK", 1},
		{"Feature", 2},
		{"TaSk", 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseTaskType(tt.input)
			assert.NoError(t, err, "ParseTaskType should not return error for: %s", tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseTaskType_Invalid(t *testing.T) {
	t.Parallel()
	tests := []string{
		"bug",
		"story",
		"epic",
		"",
		"123",
		"task ",
		" feature",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseTaskType(input)
			assert.Error(t, err)
		})
	}
}
