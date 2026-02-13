package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDoneArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          []string
		expected      *DoneInput
		expectedError string
	}{
		{
			name: "valid task ID",
			args: []string{"42"},
			expected: &DoneInput{
				TaskID: 42,
			},
		},
		{
			name: "large task ID",
			args: []string{"999999"},
			expected: &DoneInput{
				TaskID: 999999,
			},
		},
		{
			name: "task ID of 1",
			args: []string{"1"},
			expected: &DoneInput{
				TaskID: 1,
			},
		},
		{
			name: "task ID of 0",
			args: []string{"0"},
			expected: &DoneInput{
				TaskID: 0,
			},
		},
		{
			name: "negative task ID",
			args: []string{"-5"},
			expected: &DoneInput{
				TaskID: -5,
			},
		},
		{
			name:          "no arguments",
			args:          []string{},
			expectedError: "task ID is required",
		},
		{
			name:          "invalid task ID - letters",
			args:          []string{"abc"},
			expectedError: "invalid task ID: abc",
		},
		{
			name:          "invalid task ID - alphanumeric",
			args:          []string{"42abc"},
			expectedError: "invalid task ID: 42abc",
		},
		{
			name:          "invalid task ID - special characters",
			args:          []string{"#42"},
			expectedError: "invalid task ID: #42",
		},
		{
			name:          "invalid task ID - decimal",
			args:          []string{"42.5"},
			expectedError: "invalid task ID: 42.5",
		},
		{
			name:          "invalid task ID - empty string",
			args:          []string{""},
			expectedError: "invalid task ID: ",
		},
		{
			name:          "invalid task ID - whitespace",
			args:          []string{" "},
			expectedError: "invalid task ID:  ",
		},
		{
			name: "extra arguments ignored",
			args: []string{"42", "extra", "args"},
			expected: &DoneInput{
				TaskID: 42,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseDoneArgs(tt.args)

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

func TestFormatDoneOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *DoneResult
		expected string
	}{
		{
			name: "basic task move",
			result: &DoneResult{
				TaskID:     42,
				FromColumn: "In Progress",
				ToColumn:   "Done",
			},
			expected: "Task 42 moved from In Progress to Done",
		},
		{
			name: "task with single digit ID",
			result: &DoneResult{
				TaskID:     1,
				FromColumn: "To Do",
				ToColumn:   "Completed",
			},
			expected: "Task 1 moved from To Do to Completed",
		},
		{
			name: "task with large ID",
			result: &DoneResult{
				TaskID:     999999,
				FromColumn: "Review",
				ToColumn:   "Finished",
			},
			expected: "Task 999999 moved from Review to Finished",
		},
		{
			name: "column with spaces",
			result: &DoneResult{
				TaskID:     100,
				FromColumn: "Work In Progress",
				ToColumn:   "All Done",
			},
			expected: "Task 100 moved from Work In Progress to All Done",
		},
		{
			name: "column with special characters",
			result: &DoneResult{
				TaskID:     42,
				FromColumn: "In Progress",
				ToColumn:   "Done ✓",
			},
			expected: "Task 42 moved from In Progress to Done ✓",
		},
		{
			name: "column with hyphen",
			result: &DoneResult{
				TaskID:     7,
				FromColumn: "in-progress",
				ToColumn:   "done-tasks",
			},
			expected: "Task 7 moved from in-progress to done-tasks",
		},
		{
			name: "empty from column",
			result: &DoneResult{
				TaskID:     10,
				FromColumn: "",
				ToColumn:   "Done",
			},
			expected: "Task 10 moved from  to Done",
		},
		{
			name: "empty to column",
			result: &DoneResult{
				TaskID:     20,
				FromColumn: "To Do",
				ToColumn:   "",
			},
			expected: "Task 20 moved from To Do to ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatDoneOutput(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatDoneJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *DoneResult
		expected map[string]any
	}{
		{
			name: "basic task move",
			result: &DoneResult{
				TaskID:     42,
				FromColumn: "In Progress",
				ToColumn:   "Done",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     42,
				"from_column": "In Progress",
				"to_column":   "Done",
			},
		},
		{
			name: "task with single digit ID",
			result: &DoneResult{
				TaskID:     1,
				FromColumn: "To Do",
				ToColumn:   "Completed",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     1,
				"from_column": "To Do",
				"to_column":   "Completed",
			},
		},
		{
			name: "task with large ID",
			result: &DoneResult{
				TaskID:     999999,
				FromColumn: "Review",
				ToColumn:   "Finished",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     999999,
				"from_column": "Review",
				"to_column":   "Finished",
			},
		},
		{
			name: "column with spaces",
			result: &DoneResult{
				TaskID:     100,
				FromColumn: "Work In Progress",
				ToColumn:   "All Done",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     100,
				"from_column": "Work In Progress",
				"to_column":   "All Done",
			},
		},
		{
			name: "column with special characters",
			result: &DoneResult{
				TaskID:     42,
				FromColumn: "In Progress",
				ToColumn:   "Done ✓",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     42,
				"from_column": "In Progress",
				"to_column":   "Done ✓",
			},
		},
		{
			name: "column with hyphen",
			result: &DoneResult{
				TaskID:     7,
				FromColumn: "in-progress",
				ToColumn:   "done-tasks",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     7,
				"from_column": "in-progress",
				"to_column":   "done-tasks",
			},
		},
		{
			name: "empty from column",
			result: &DoneResult{
				TaskID:     10,
				FromColumn: "",
				ToColumn:   "Done",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     10,
				"from_column": "",
				"to_column":   "Done",
			},
		},
		{
			name: "empty to column",
			result: &DoneResult{
				TaskID:     20,
				FromColumn: "To Do",
				ToColumn:   "",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     20,
				"from_column": "To Do",
				"to_column":   "",
			},
		},
		{
			name: "both columns empty",
			result: &DoneResult{
				TaskID:     30,
				FromColumn: "",
				ToColumn:   "",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     30,
				"from_column": "",
				"to_column":   "",
			},
		},
		{
			name: "zero task ID",
			result: &DoneResult{
				TaskID:     0,
				FromColumn: "To Do",
				ToColumn:   "Done",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     0,
				"from_column": "To Do",
				"to_column":   "Done",
			},
		},
		{
			name: "negative task ID",
			result: &DoneResult{
				TaskID:     -1,
				FromColumn: "To Do",
				ToColumn:   "Done",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     -1,
				"from_column": "To Do",
				"to_column":   "Done",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatDoneJSON(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}
