package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDeleteArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          []string
		expected      *DeleteInput
		expectedError string
	}{
		{
			name: "valid task ID",
			args: []string{"42"},
			expected: &DeleteInput{
				TaskID: 42,
			},
		},
		{
			name: "large task ID",
			args: []string{"999999"},
			expected: &DeleteInput{
				TaskID: 999999,
			},
		},
		{
			name: "task ID of 1",
			args: []string{"1"},
			expected: &DeleteInput{
				TaskID: 1,
			},
		},
		{
			name: "task ID of 0",
			args: []string{"0"},
			expected: &DeleteInput{
				TaskID: 0,
			},
		},
		{
			name: "negative task ID",
			args: []string{"-5"},
			expected: &DeleteInput{
				TaskID: -5,
			},
		},
		{
			name:          "no arguments",
			args:          []string{},
			expectedError: "task ID is required",
		},
		{
			name:          "invalid task ID - not a number",
			args:          []string{"not-a-number"},
			expectedError: "invalid task ID: not-a-number",
		},
		{
			name:          "invalid task ID - alphanumeric",
			args:          []string{"abc123"},
			expectedError: "invalid task ID: abc123",
		},
		{
			name:          "invalid task ID - float",
			args:          []string{"42.5"},
			expectedError: "invalid task ID: 42.5",
		},
		{
			name:          "invalid task ID - special characters",
			args:          []string{"#42"},
			expectedError: "invalid task ID: #42",
		},
		{
			name: "extra arguments ignored",
			args: []string{"42", "extra", "args"},
			expected: &DeleteInput{
				TaskID: 42,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseDeleteArgs(tt.args)

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

func TestFormatDeleteOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *DeleteResult
		expected string
	}{
		{
			name: "task ID 42",
			result: &DeleteResult{
				TaskID: 42,
			},
			expected: "Task 42 deleted successfully",
		},
		{
			name: "task ID 1",
			result: &DeleteResult{
				TaskID: 1,
			},
			expected: "Task 1 deleted successfully",
		},
		{
			name: "large task ID",
			result: &DeleteResult{
				TaskID: 999999,
			},
			expected: "Task 999999 deleted successfully",
		},
		{
			name: "task ID 0",
			result: &DeleteResult{
				TaskID: 0,
			},
			expected: "Task 0 deleted successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatDeleteOutput(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatDeleteJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *DeleteResult
		expected map[string]any
	}{
		{
			name: "task ID 42",
			result: &DeleteResult{
				TaskID: 42,
			},
			expected: map[string]any{
				"success": true,
				"task_id": 42,
			},
		},
		{
			name: "task ID 1",
			result: &DeleteResult{
				TaskID: 1,
			},
			expected: map[string]any{
				"success": true,
				"task_id": 1,
			},
		},
		{
			name: "large task ID",
			result: &DeleteResult{
				TaskID: 999999,
			},
			expected: map[string]any{
				"success": true,
				"task_id": 999999,
			},
		},
		{
			name: "task ID 0",
			result: &DeleteResult{
				TaskID: 0,
			},
			expected: map[string]any{
				"success": true,
				"task_id": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatDeleteJSON(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatConfirmationPrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		taskID    int
		taskTitle string
		expected  string
	}{
		{
			name:      "simple task",
			taskID:    42,
			taskTitle: "Fix bug in login",
			expected:  "Delete task #42: 'Fix bug in login'? (y/N): ",
		},
		{
			name:      "task with special characters",
			taskID:    100,
			taskTitle: "Update API endpoint: /users/{id}",
			expected:  "Delete task #100: 'Update API endpoint: /users/{id}'? (y/N): ",
		},
		{
			name:      "empty task title",
			taskID:    1,
			taskTitle: "",
			expected:  "Delete task #1: ''? (y/N): ",
		},
		{
			name:      "task with quotes",
			taskID:    7,
			taskTitle: "Implement \"dark mode\" feature",
			expected:  "Delete task #7: 'Implement \"dark mode\" feature'? (y/N): ",
		},
		{
			name:      "long task title",
			taskID:    999,
			taskTitle: "This is a very long task title that contains a lot of information and details about what needs to be done",
			expected:  "Delete task #999: 'This is a very long task title that contains a lot of information and details about what needs to be done'? (y/N): ",
		},
		{
			name:      "task with newline characters in title",
			taskID:    5,
			taskTitle: "Task with\nnewline",
			expected:  "Delete task #5: 'Task with\nnewline'? (y/N): ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatConfirmationPrompt(tt.taskID, tt.taskTitle)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestIsConfirmationYes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		response string
		expected bool
	}{
		{
			name:     "lowercase y",
			response: "y",
			expected: true,
		},
		{
			name:     "uppercase Y",
			response: "Y",
			expected: true,
		},
		{
			name:     "lowercase yes",
			response: "yes",
			expected: true,
		},
		{
			name:     "uppercase YES",
			response: "YES",
			expected: true,
		},
		{
			name:     "mixed case Yes",
			response: "Yes",
			expected: true,
		},
		{
			name:     "mixed case YeS",
			response: "YeS",
			expected: true,
		},
		{
			name:     "y with leading whitespace",
			response: "  y",
			expected: true,
		},
		{
			name:     "y with trailing whitespace",
			response: "y  ",
			expected: true,
		},
		{
			name:     "yes with surrounding whitespace",
			response: "  yes  ",
			expected: true,
		},
		{
			name:     "lowercase n",
			response: "n",
			expected: false,
		},
		{
			name:     "uppercase N",
			response: "N",
			expected: false,
		},
		{
			name:     "lowercase no",
			response: "no",
			expected: false,
		},
		{
			name:     "uppercase NO",
			response: "NO",
			expected: false,
		},
		{
			name:     "empty string",
			response: "",
			expected: false,
		},
		{
			name:     "whitespace only",
			response: "   ",
			expected: false,
		},
		{
			name:     "random text",
			response: "maybe",
			expected: false,
		},
		{
			name:     "number",
			response: "1",
			expected: false,
		},
		{
			name:     "partial yes",
			response: "ye",
			expected: false,
		},
		{
			name:     "yes with extra text",
			response: "yes please",
			expected: false,
		},
		{
			name:     "y with extra text",
			response: "yeah",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsConfirmationYes(tt.response)
			assert.Equal(t, tt.expected, result)
		})
	}
}
