package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAssignArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          []string
		expected      *AssignInput
		expectedError string
	}{
		{
			name: "task ID only",
			args: []string{"42"},
			expected: &AssignInput{
				TaskID:       42,
				AssigneeName: "",
			},
		},
		{
			name: "task ID with assignee name",
			args: []string{"42", "alice"},
			expected: &AssignInput{
				TaskID:       42,
				AssigneeName: "alice",
			},
		},
		{
			name:          "no arguments",
			args:          []string{},
			expectedError: "task ID is required",
		},
		{
			name:          "invalid task ID",
			args:          []string{"not-a-number"},
			expectedError: "invalid task ID: not-a-number",
		},
		{
			name: "negative task ID",
			args: []string{"-5"},
			expected: &AssignInput{
				TaskID:       -5,
				AssigneeName: "",
			},
		},
		{
			name: "task ID with multi-word assignee name (takes first word only)",
			args: []string{"100", "alice smith"},
			expected: &AssignInput{
				TaskID:       100,
				AssigneeName: "alice smith",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseAssignArgs(tt.args)

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

func TestFormatAssignOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *AssignResult
		expected string
	}{
		{
			name: "assign to user",
			result: &AssignResult{
				TaskID:       42,
				AssigneeName: "alice",
				Cleared:      false,
			},
			expected: "Task 42 assigned to @alice",
		},
		{
			name: "clear assignee",
			result: &AssignResult{
				TaskID:  100,
				Cleared: true,
			},
			expected: "Task 100 assignee cleared",
		},
		{
			name: "assign to user with hyphenated name",
			result: &AssignResult{
				TaskID:       7,
				AssigneeName: "alice-bob",
				Cleared:      false,
			},
			expected: "Task 7 assigned to @alice-bob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatAssignOutput(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatAssignJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *AssignResult
		expected map[string]any
	}{
		{
			name: "assign to user",
			result: &AssignResult{
				TaskID:       42,
				AssigneeName: "alice",
				Cleared:      false,
			},
			expected: map[string]any{
				"success":  true,
				"task_id":  42,
				"assignee": "alice",
			},
		},
		{
			name: "clear assignee",
			result: &AssignResult{
				TaskID:  100,
				Cleared: true,
			},
			expected: map[string]any{
				"success":  true,
				"task_id":  100,
				"assignee": nil,
			},
		},
		{
			name: "assign to user with special characters",
			result: &AssignResult{
				TaskID:       999,
				AssigneeName: "user.name@domain",
				Cleared:      false,
			},
			expected: map[string]any{
				"success":  true,
				"task_id":  999,
				"assignee": "user.name@domain",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatAssignJSON(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}
