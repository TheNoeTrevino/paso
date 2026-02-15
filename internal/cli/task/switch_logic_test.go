package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSwitchArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          []string
		expected      *SwitchInput
		expectedError string
	}{
		{
			name: "valid task ID and project ID",
			args: []string{"42", "5"},
			expected: &SwitchInput{
				TaskID:    42,
				ProjectID: 5,
			},
		},
		{
			name:          "no arguments",
			args:          []string{},
			expectedError: "task ID and project ID are required",
		},
		{
			name:          "only one argument",
			args:          []string{"42"},
			expectedError: "task ID and project ID are required",
		},
		{
			name:          "too many arguments",
			args:          []string{"42", "5", "extra"},
			expectedError: "expected 2 arguments, got 3",
		},
		{
			name:          "non-integer task ID",
			args:          []string{"not-a-number", "5"},
			expectedError: "invalid task ID: not-a-number",
		},
		{
			name:          "non-integer project ID",
			args:          []string{"42", "not-a-number"},
			expectedError: "invalid project ID: not-a-number",
		},
		{
			name:          "zero task ID",
			args:          []string{"0", "5"},
			expectedError: "task ID must be a positive integer",
		},
		{
			name:          "zero project ID",
			args:          []string{"42", "0"},
			expectedError: "project ID must be a positive integer",
		},
		{
			name:          "negative task ID",
			args:          []string{"-1", "5"},
			expectedError: "task ID must be a positive integer",
		},
		{
			name:          "negative project ID",
			args:          []string{"42", "-3"},
			expectedError: "project ID must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseSwitchArgs(tt.args)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFormatSwitchOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *SwitchResult
		expected string
	}{
		{
			name: "move to project",
			result: &SwitchResult{
				TaskID:      42,
				ProjectID:   5,
				ProjectName: "Backend",
			},
			expected: "Task 42 successfully moved to project Backend",
		},
		{
			name: "move to project with spaces in name",
			result: &SwitchResult{
				TaskID:      7,
				ProjectID:   10,
				ProjectName: "My Project",
			},
			expected: "Task 7 successfully moved to project My Project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatSwitchOutput(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatSwitchJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *SwitchResult
		expected map[string]any
	}{
		{
			name: "basic result",
			result: &SwitchResult{
				TaskID:      42,
				ProjectID:   5,
				ProjectName: "Backend",
			},
			expected: map[string]any{
				"task_id":      42,
				"project_id":   5,
				"project_name": "Backend",
			},
		},
		{
			name: "result with special characters in name",
			result: &SwitchResult{
				TaskID:      999,
				ProjectID:   1,
				ProjectName: "Q1-2026 Release",
			},
			expected: map[string]any{
				"task_id":      999,
				"project_id":   1,
				"project_name": "Q1-2026 Release",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatSwitchJSON(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}
