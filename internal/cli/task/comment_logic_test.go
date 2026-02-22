package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/config/colors"
)

func TestValidateCommentMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		message       string
		expectedError string
	}{
		{
			name:          "valid message - short",
			message:       "This is a valid comment",
			expectedError: "",
		},
		{
			name:          "valid message - empty",
			message:       "",
			expectedError: "",
		},
		{
			name:          "valid message - exactly 1000 chars",
			message:       string(make([]byte, 1000)),
			expectedError: "",
		},
		{
			name:          "invalid message - 1001 chars",
			message:       string(make([]byte, 1001)),
			expectedError: "message exceeds 1000 character limit",
		},
		{
			name:          "invalid message - very long",
			message:       string(make([]byte, 5000)),
			expectedError: "message exceeds 1000 character limit",
		},
		{
			name:          "valid message - unicode characters",
			message:       "Hello 世界 🌍",
			expectedError: "",
		},
		{
			name:          "valid message - special characters",
			message:       "Test with special chars: @#$%^&*()_+-=[]{}|;':\",./<>?",
			expectedError: "",
		},
		{
			name:          "valid message - newlines and tabs",
			message:       "Line 1\nLine 2\tTabbed",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateCommentMessage(tt.message)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFormatCommentQuiet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		commentID int
		expected  string
	}{
		{
			name:      "positive comment ID",
			commentID: 42,
			expected:  "42\n",
		},
		{
			name:      "zero comment ID",
			commentID: 0,
			expected:  "0\n",
		},
		{
			name:      "large comment ID",
			commentID: 999999,
			expected:  "999999\n",
		},
		{
			name:      "single digit",
			commentID: 1,
			expected:  "1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatCommentQuiet(tt.commentID)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatCommentJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *CommentResult
		expected map[string]any
	}{
		{
			name: "complete comment result",
			result: &CommentResult{
				CommentID:    123,
				TaskID:       42,
				Message:      "This is a test comment",
				Author:       "alice",
				CreatedAt:    "2024-01-15 10:30:00",
				TaskTitle:    "Fix bug in login",
				TaskNumber: 100,
				ProjectName:  "Auth Service",
			},
			expected: map[string]any{
				"success": true,
				"comment": map[string]any{
					"id":         123,
					"task_id":    42,
					"message":    "This is a test comment",
					"author":     "alice",
					"created_at": "2024-01-15 10:30:00",
				},
				"task": map[string]any{
					"id":            42,
					"title":         "Fix bug in login",
					"task_number": 100,
					"project":       "Auth Service",
				},
			},
		},
		{
			name: "comment with empty message",
			result: &CommentResult{
				CommentID:    1,
				TaskID:       1,
				Message:      "",
				Author:       "bob",
				CreatedAt:    "2024-01-01 00:00:00",
				TaskTitle:    "Task",
				TaskNumber: 1,
				ProjectName:  "Project",
			},
			expected: map[string]any{
				"success": true,
				"comment": map[string]any{
					"id":         1,
					"task_id":    1,
					"message":    "",
					"author":     "bob",
					"created_at": "2024-01-01 00:00:00",
				},
				"task": map[string]any{
					"id":            1,
					"title":         "Task",
					"task_number": 1,
					"project":       "Project",
				},
			},
		},
		{
			name: "comment with special characters",
			result: &CommentResult{
				CommentID:    999,
				TaskID:       555,
				Message:      "Test with \"quotes\" and 'apostrophes' & symbols",
				Author:       "user.name@domain",
				CreatedAt:    "2024-12-31 23:59:59",
				TaskTitle:    "Task with symbols: @#$%",
				TaskNumber: 777,
				ProjectName:  "Project-Name_123",
			},
			expected: map[string]any{
				"success": true,
				"comment": map[string]any{
					"id":         999,
					"task_id":    555,
					"message":    "Test with \"quotes\" and 'apostrophes' & symbols",
					"author":     "user.name@domain",
					"created_at": "2024-12-31 23:59:59",
				},
				"task": map[string]any{
					"id":            555,
					"title":         "Task with symbols: @#$%",
					"task_number": 777,
					"project":       "Project-Name_123",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatCommentJSON(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatCommentHuman(t *testing.T) {
	t.Parallel()

	// Create test color schemes
	testColorScheme := colors.ColorScheme{
		Normal: "#FFFFFF",
		Create: "#00FF00",
		Delete: "#FF0000",
	}

	tests := []struct {
		name             string
		result           *CommentResult
		colorScheme      colors.ColorScheme
		expectContains   []string
		expectNotContain []string
	}{
		{
			name: "standard comment output",
			result: &CommentResult{
				CommentID:    42,
				TaskID:       100,
				Message:      "This is a test comment",
				Author:       "alice",
				CreatedAt:    "2024-01-15 10:30:00",
				TaskTitle:    "Fix bug",
				TaskNumber: 123,
				ProjectName:  "Test Project",
			},
			colorScheme: testColorScheme,
			expectContains: []string{
				"Comment added",
				"#123",
				"Fix bug",
				"Test Project",
				"This is a test comment",
				"42",
			},
		},
		{
			name: "long message gets truncated",
			result: &CommentResult{
				CommentID:    1,
				TaskID:       1,
				Message:      "This is a very long message that should be truncated because it exceeds the 60 character limit for display purposes",
				Author:       "bob",
				CreatedAt:    "2024-01-01 00:00:00",
				TaskTitle:    "Task",
				TaskNumber: 1,
				ProjectName:  "Project",
			},
			colorScheme: testColorScheme,
			expectContains: []string{
				"Comment added",
				"...", // Truncation indicator
			},
			expectNotContain: []string{
				"for display purposes", // End of the long message should be truncated
			},
		},
		{
			name: "message exactly 60 chars - no truncation",
			result: &CommentResult{
				CommentID:    2,
				TaskID:       2,
				Message:      "This message is exactly sixty characters long for the test",
				Author:       "charlie",
				CreatedAt:    "2024-02-01 12:00:00",
				TaskTitle:    "Another Task",
				TaskNumber: 50,
				ProjectName:  "Another Project",
			},
			colorScheme: testColorScheme,
			expectContains: []string{
				"This message is exactly sixty characters long for the test",
			},
			expectNotContain: []string{
				"...",
			},
		},
		{
			name: "comment with special characters in task title",
			result: &CommentResult{
				CommentID:    999,
				TaskID:       888,
				Message:      "Comment message",
				Author:       "dave",
				CreatedAt:    "2024-03-01 08:00:00",
				TaskTitle:    "Task with @mentions and #hashtags",
				TaskNumber: 777,
				ProjectName:  "Special-Project_123",
			},
			colorScheme: testColorScheme,
			expectContains: []string{
				"Task with @mentions and #hashtags",
				"Special-Project_123",
				"#777",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatCommentHuman(tt.result, tt.colorScheme)

			for _, expected := range tt.expectContains {
				assert.Contains(t, output, expected, "Expected output to contain: %s", expected)
			}

			for _, notExpected := range tt.expectNotContain {
				assert.NotContains(t, output, notExpected, "Expected output to NOT contain: %s", notExpected)
			}

			// Verify output is not empty
			assert.NotEmpty(t, output)
		})
	}
}

func TestFormatCommentHuman_StructureValidation(t *testing.T) {
	t.Parallel()

	testColorScheme := colors.ColorScheme{
		Normal: "#FFFFFF",
		Create: "#00FF00",
		Delete: "#FF0000",
	}

	result := &CommentResult{
		CommentID:    42,
		TaskID:       100,
		Message:      "Test message",
		Author:       "alice",
		CreatedAt:    "2024-01-15 10:30:00",
		TaskTitle:    "Fix bug",
		TaskNumber: 123,
		ProjectName:  "Test Project",
	}

	output := FormatCommentHuman(result, testColorScheme)

	// Verify the output contains all detail keys
	assert.Contains(t, output, "Task:")
	assert.Contains(t, output, "Project:")
	assert.Contains(t, output, "Message:")
	assert.Contains(t, output, "Comment ID:")

	// Verify it contains the checkmark symbol (success indicator)
	assert.Contains(t, output, "✓")

	// Verify it ends with newline
	assert.True(t, len(output) > 0 && output[len(output)-1] == '\n', "Output should end with newline")
}

func TestCommentInput_Struct(t *testing.T) {
	t.Parallel()

	// Test struct creation and field access
	input := &CommentInput{
		TaskID:  42,
		Message: "Test message",
		Author:  "alice",
	}

	assert.Equal(t, 42, input.TaskID)
	assert.Equal(t, "Test message", input.Message)
	assert.Equal(t, "alice", input.Author)
}

func TestCommentResult_Struct(t *testing.T) {
	t.Parallel()

	// Test struct creation and field access
	result := &CommentResult{
		CommentID:    123,
		TaskID:       42,
		Message:      "Test message",
		Author:       "alice",
		CreatedAt:    "2024-01-15 10:30:00",
		TaskTitle:    "Fix bug",
		TaskNumber: 100,
		ProjectName:  "Test Project",
	}

	assert.Equal(t, 123, result.CommentID)
	assert.Equal(t, 42, result.TaskID)
	assert.Equal(t, "Test message", result.Message)
	assert.Equal(t, "alice", result.Author)
	assert.Equal(t, "2024-01-15 10:30:00", result.CreatedAt)
	assert.Equal(t, "Fix bug", result.TaskTitle)
	assert.Equal(t, 100, result.TaskNumber)
	assert.Equal(t, "Test Project", result.ProjectName)
}
