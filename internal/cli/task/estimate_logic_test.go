package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEstimateArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          []string
		clearFlag     bool
		expected      *EstimateInput
		expectedError string
	}{
		{
			name:      "task ID with estimate",
			args:      []string{"42", "2d"},
			clearFlag: false,
			expected: &EstimateInput{
				TaskID:   42,
				Estimate: "2d",
				Clear:    false,
			},
		},
		{
			name:      "task ID only (no estimate)",
			args:      []string{"42"},
			clearFlag: false,
			expected: &EstimateInput{
				TaskID:   42,
				Estimate: "",
				Clear:    false,
			},
		},
		{
			name:      "task ID with clear flag",
			args:      []string{"42"},
			clearFlag: true,
			expected: &EstimateInput{
				TaskID:   42,
				Estimate: "",
				Clear:    true,
			},
		},
		{
			name:          "no arguments",
			args:          []string{},
			clearFlag:     false,
			expectedError: "task ID is required",
		},
		{
			name:          "invalid task ID",
			args:          []string{"abc"},
			clearFlag:     false,
			expectedError: "invalid task ID: abc",
		},
		{
			name:      "complex estimate format",
			args:      []string{"100", "1w2d3h"},
			clearFlag: false,
			expected: &EstimateInput{
				TaskID:   100,
				Estimate: "1w2d3h",
				Clear:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseEstimateArgs(tt.args, tt.clearFlag)

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

func TestFormatEstimateOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *EstimateResult
		expected string
	}{
		{
			name: "set estimate",
			result: &EstimateResult{
				TaskID:   42,
				Estimate: "2d",
				Cleared:  false,
			},
			expected: "Task 42 estimate set to 2d",
		},
		{
			name: "clear estimate",
			result: &EstimateResult{
				TaskID:  100,
				Cleared: true,
			},
			expected: "Task 100 estimate cleared",
		},
		{
			name: "complex estimate",
			result: &EstimateResult{
				TaskID:   7,
				Estimate: "1w2d3h",
				Cleared:  false,
			},
			expected: "Task 7 estimate set to 1w2d3h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatEstimateOutput(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatEstimateJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *EstimateResult
		expected map[string]any
	}{
		{
			name: "set estimate",
			result: &EstimateResult{
				TaskID:   42,
				Estimate: "2d",
				Cleared:  false,
			},
			expected: map[string]any{
				"success":  true,
				"task_id":  42,
				"estimate": "2d",
			},
		},
		{
			name: "clear estimate",
			result: &EstimateResult{
				TaskID:  100,
				Cleared: true,
			},
			expected: map[string]any{
				"success":  true,
				"task_id":  100,
				"estimate": nil,
			},
		},
		{
			name: "complex estimate",
			result: &EstimateResult{
				TaskID:   999,
				Estimate: "3m2w1d",
				Cleared:  false,
			},
			expected: map[string]any{
				"success":  true,
				"task_id":  999,
				"estimate": "3m2w1d",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatEstimateJSON(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}
