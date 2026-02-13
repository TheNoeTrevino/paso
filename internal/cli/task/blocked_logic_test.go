package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestFilterBlockedTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		tasksByColumn map[int][]*models.TaskSummary
		expected      []*models.TaskSummary
	}{
		{
			name: "filters blocked tasks from multiple columns",
			tasksByColumn: map[int][]*models.TaskSummary{
				1: {
					{ID: 1, Title: "Task 1", IsBlocked: false},
					{ID: 2, Title: "Task 2", IsBlocked: true},
				},
				2: {
					{ID: 3, Title: "Task 3", IsBlocked: true},
					{ID: 4, Title: "Task 4", IsBlocked: false},
				},
			},
			expected: []*models.TaskSummary{
				{ID: 2, Title: "Task 2", IsBlocked: true},
				{ID: 3, Title: "Task 3", IsBlocked: true},
			},
		},
		{
			name: "returns empty slice when no blocked tasks",
			tasksByColumn: map[int][]*models.TaskSummary{
				1: {
					{ID: 1, Title: "Task 1", IsBlocked: false},
					{ID: 2, Title: "Task 2", IsBlocked: false},
				},
			},
			expected: []*models.TaskSummary{},
		},
		{
			name: "returns all tasks when all are blocked",
			tasksByColumn: map[int][]*models.TaskSummary{
				1: {
					{ID: 1, Title: "Task 1", IsBlocked: true},
					{ID: 2, Title: "Task 2", IsBlocked: true},
				},
			},
			expected: []*models.TaskSummary{
				{ID: 1, Title: "Task 1", IsBlocked: true},
				{ID: 2, Title: "Task 2", IsBlocked: true},
			},
		},
		{
			name:          "returns empty slice when input is empty",
			tasksByColumn: map[int][]*models.TaskSummary{},
			expected:      []*models.TaskSummary{},
		},
		{
			name: "handles columns with no tasks",
			tasksByColumn: map[int][]*models.TaskSummary{
				1: {},
				2: {
					{ID: 5, Title: "Task 5", IsBlocked: true},
				},
			},
			expected: []*models.TaskSummary{
				{ID: 5, Title: "Task 5", IsBlocked: true},
			},
		},
		{
			name:          "handles nil map",
			tasksByColumn: nil,
			expected:      []*models.TaskSummary{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FilterBlockedTasks(tt.tasksByColumn)

			if len(tt.expected) == 0 {
				assert.Empty(t, result)
			} else {
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestFormatBlockedOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *BlockedResult
		expected string
	}{
		{
			name: "formats single blocked task with default priority",
			result: &BlockedResult{
				Tasks: []*models.TaskSummary{
					{ID: 42, Title: "Fix bug", PriorityDescription: "medium"},
				},
				Count: 1,
			},
			expected: "Found 1 blocked tasks:\n\n  [42] Fix bug (BLOCKED)\n",
		},
		{
			name: "formats single blocked task with high priority",
			result: &BlockedResult{
				Tasks: []*models.TaskSummary{
					{ID: 42, Title: "Fix bug", PriorityDescription: "high"},
				},
				Count: 1,
			},
			expected: "Found 1 blocked tasks:\n\n  [42] Fix bug [high] (BLOCKED)\n",
		},
		{
			name: "formats multiple blocked tasks",
			result: &BlockedResult{
				Tasks: []*models.TaskSummary{
					{ID: 1, Title: "Task 1", PriorityDescription: "low"},
					{ID: 2, Title: "Task 2", PriorityDescription: "medium"},
					{ID: 3, Title: "Task 3", PriorityDescription: "high"},
				},
				Count: 3,
			},
			expected: "Found 3 blocked tasks:\n\n  [1] Task 1 [low] (BLOCKED)\n  [2] Task 2 (BLOCKED)\n  [3] Task 3 [high] (BLOCKED)\n",
		},
		{
			name: "formats zero blocked tasks",
			result: &BlockedResult{
				Tasks: []*models.TaskSummary{},
				Count: 0,
			},
			expected: "No blocked tasks found",
		},
		{
			name: "formats task with empty priority description",
			result: &BlockedResult{
				Tasks: []*models.TaskSummary{
					{ID: 100, Title: "No priority", PriorityDescription: ""},
				},
				Count: 1,
			},
			expected: "Found 1 blocked tasks:\n\n  [100] No priority (BLOCKED)\n",
		},
		{
			name: "formats task with special characters in title",
			result: &BlockedResult{
				Tasks: []*models.TaskSummary{
					{ID: 7, Title: "Task with \"quotes\" & symbols", PriorityDescription: "high"},
				},
				Count: 1,
			},
			expected: "Found 1 blocked tasks:\n\n  [7] Task with \"quotes\" & symbols [high] (BLOCKED)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatBlockedOutput(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatBlockedJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *BlockedResult
		expected map[string]any
	}{
		{
			name: "formats single blocked task",
			result: &BlockedResult{
				Tasks: []*models.TaskSummary{
					{ID: 42, Title: "Fix bug"},
				},
				Count: 1,
			},
			expected: map[string]any{
				"success": true,
				"tasks": []*models.TaskSummary{
					{ID: 42, Title: "Fix bug"},
				},
				"count": 1,
			},
		},
		{
			name: "formats multiple blocked tasks",
			result: &BlockedResult{
				Tasks: []*models.TaskSummary{
					{ID: 1, Title: "Task 1"},
					{ID: 2, Title: "Task 2"},
				},
				Count: 2,
			},
			expected: map[string]any{
				"success": true,
				"tasks": []*models.TaskSummary{
					{ID: 1, Title: "Task 1"},
					{ID: 2, Title: "Task 2"},
				},
				"count": 2,
			},
		},
		{
			name: "formats zero blocked tasks",
			result: &BlockedResult{
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
			name: "formats with nil tasks slice",
			result: &BlockedResult{
				Tasks: nil,
				Count: 0,
			},
			expected: map[string]any{
				"success": true,
				"tasks":   ([]*models.TaskSummary)(nil),
				"count":   0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatBlockedJSON(tt.result)
			assert.Equal(t, tt.expected["success"], output["success"])
			assert.Equal(t, tt.expected["count"], output["count"])
			assert.Equal(t, tt.expected["tasks"], output["tasks"])
		})
	}
}

func TestFormatBlockedQuiet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *BlockedResult
		expected string
	}{
		{
			name: "formats single blocked task ID",
			result: &BlockedResult{
				Tasks: []*models.TaskSummary{
					{ID: 42, Title: "Fix bug"},
				},
				Count: 1,
			},
			expected: "42\n",
		},
		{
			name: "formats multiple blocked task IDs",
			result: &BlockedResult{
				Tasks: []*models.TaskSummary{
					{ID: 1, Title: "Task 1"},
					{ID: 2, Title: "Task 2"},
					{ID: 3, Title: "Task 3"},
				},
				Count: 3,
			},
			expected: "1\n2\n3\n",
		},
		{
			name: "formats zero blocked tasks",
			result: &BlockedResult{
				Tasks: []*models.TaskSummary{},
				Count: 0,
			},
			expected: "",
		},
		{
			name: "formats with nil tasks slice",
			result: &BlockedResult{
				Tasks: nil,
				Count: 0,
			},
			expected: "",
		},
		{
			name: "formats large task IDs",
			result: &BlockedResult{
				Tasks: []*models.TaskSummary{
					{ID: 999999, Title: "Large ID"},
				},
				Count: 1,
			},
			expected: "999999\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatBlockedQuiet(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}
