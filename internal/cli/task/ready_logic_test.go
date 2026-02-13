package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestFormatReadyOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *ReadyResult
		expected string
	}{
		{
			name: "no tasks",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{},
				Count: 0,
			},
			expected: "No ready tasks found",
		},
		{
			name: "single task with no priority",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{
					{
						ID:                  42,
						Title:               "Fix bug",
						PriorityDescription: "",
					},
				},
				Count: 1,
			},
			expected: "Found 1 ready tasks:\n\n  [42] Fix bug\n",
		},
		{
			name: "single task with medium priority (should not show)",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{
					{
						ID:                  42,
						Title:               "Fix bug",
						PriorityDescription: "medium",
					},
				},
				Count: 1,
			},
			expected: "Found 1 ready tasks:\n\n  [42] Fix bug\n",
		},
		{
			name: "single task with high priority",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{
					{
						ID:                  42,
						Title:               "Fix critical bug",
						PriorityDescription: "high",
					},
				},
				Count: 1,
			},
			expected: "Found 1 ready tasks:\n\n  [42] Fix critical bug [high]\n",
		},
		{
			name: "single task with low priority",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{
					{
						ID:                  100,
						Title:               "Nice to have feature",
						PriorityDescription: "low",
					},
				},
				Count: 1,
			},
			expected: "Found 1 ready tasks:\n\n  [100] Nice to have feature [low]\n",
		},
		{
			name: "multiple tasks with mixed priorities",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{
					{
						ID:                  1,
						Title:               "Critical bug",
						PriorityDescription: "high",
					},
					{
						ID:                  2,
						Title:               "Regular task",
						PriorityDescription: "medium",
					},
					{
						ID:                  3,
						Title:               "Low priority task",
						PriorityDescription: "low",
					},
				},
				Count: 3,
			},
			expected: "Found 3 ready tasks:\n\n  [1] Critical bug [high]\n  [2] Regular task\n  [3] Low priority task [low]\n",
		},
		{
			name: "task with empty priority description",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{
					{
						ID:                  99,
						Title:               "Task without priority",
						PriorityDescription: "",
					},
				},
				Count: 1,
			},
			expected: "Found 1 ready tasks:\n\n  [99] Task without priority\n",
		},
		{
			name: "task with special characters in title",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{
					{
						ID:                  50,
						Title:               "Fix: bug with @mentions & #tags",
						PriorityDescription: "",
					},
				},
				Count: 1,
			},
			expected: "Found 1 ready tasks:\n\n  [50] Fix: bug with @mentions & #tags\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatReadyOutput(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatReadyJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *ReadyResult
		expected map[string]any
	}{
		{
			name: "no tasks",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{},
				Count: 0,
			},
			expected: map[string]any{
				"success": true,
				"tasks":   []*models.TaskSummary{},
				"count":   0,
			},
		},
		{
			name: "single task",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{
					{
						ID:                  42,
						Title:               "Fix bug",
						PriorityDescription: "high",
					},
				},
				Count: 1,
			},
			expected: map[string]any{
				"success": true,
				"tasks": []*models.TaskSummary{
					{
						ID:                  42,
						Title:               "Fix bug",
						PriorityDescription: "high",
					},
				},
				"count": 1,
			},
		},
		{
			name: "multiple tasks",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{
					{
						ID:                  1,
						Title:               "Task 1",
						PriorityDescription: "high",
					},
					{
						ID:                  2,
						Title:               "Task 2",
						PriorityDescription: "medium",
					},
					{
						ID:                  3,
						Title:               "Task 3",
						PriorityDescription: "low",
					},
				},
				Count: 3,
			},
			expected: map[string]any{
				"success": true,
				"tasks": []*models.TaskSummary{
					{
						ID:                  1,
						Title:               "Task 1",
						PriorityDescription: "high",
					},
					{
						ID:                  2,
						Title:               "Task 2",
						PriorityDescription: "medium",
					},
					{
						ID:                  3,
						Title:               "Task 3",
						PriorityDescription: "low",
					},
				},
				"count": 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatReadyJSON(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatReadyQuiet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *ReadyResult
		expected string
	}{
		{
			name: "no tasks",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{},
				Count: 0,
			},
			expected: "",
		},
		{
			name: "single task",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{
					{
						ID:    42,
						Title: "Fix bug",
					},
				},
				Count: 1,
			},
			expected: "42\n",
		},
		{
			name: "multiple tasks",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{
					{
						ID:    1,
						Title: "Task 1",
					},
					{
						ID:    2,
						Title: "Task 2",
					},
					{
						ID:    100,
						Title: "Task 100",
					},
				},
				Count: 3,
			},
			expected: "1\n2\n100\n",
		},
		{
			name: "large task IDs",
			result: &ReadyResult{
				Tasks: []*models.TaskSummary{
					{
						ID:    999999,
						Title: "Task with large ID",
					},
				},
				Count: 1,
			},
			expected: "999999\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatReadyQuiet(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}
