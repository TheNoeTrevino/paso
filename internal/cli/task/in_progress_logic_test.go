package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInProgressArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          []string
		projectID     int
		expected      *InProgressInput
		expectedError string
	}{
		{
			name:      "list mode with project ID",
			args:      []string{},
			projectID: 1,
			expected: &InProgressInput{
				Mode:      InProgressModeList,
				ProjectID: 1,
			},
		},
		{
			name:      "list mode with project ID and extra args (ignored)",
			args:      []string{"42"},
			projectID: 5,
			expected: &InProgressInput{
				Mode:      InProgressModeList,
				ProjectID: 5,
			},
		},
		{
			name:      "move mode with task ID",
			args:      []string{"42"},
			projectID: 0,
			expected: &InProgressInput{
				Mode:   InProgressModeMove,
				TaskID: 42,
			},
		},
		{
			name:      "move mode with large task ID",
			args:      []string{"999999"},
			projectID: 0,
			expected: &InProgressInput{
				Mode:   InProgressModeMove,
				TaskID: 999999,
			},
		},
		{
			name:          "no task ID and no project ID",
			args:          []string{},
			projectID:     0,
			expectedError: "task ID is required (or use --project flag to list tasks)",
		},
		{
			name:          "invalid task ID",
			args:          []string{"abc"},
			projectID:     0,
			expectedError: "invalid task ID 'abc': must be a number",
		},
		{
			name:          "invalid task ID with special characters",
			args:          []string{"42@#$"},
			projectID:     0,
			expectedError: "invalid task ID '42@#$': must be a number",
		},
		{
			name:          "negative task ID in string",
			args:          []string{"-5"},
			projectID:     0,
			expectedError: "task ID must be a positive integer",
		},
		{
			name:          "empty string as task ID",
			args:          []string{""},
			projectID:     0,
			expectedError: "invalid task ID '': must be a number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseInProgressArgs(tt.args, tt.projectID)

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

func TestFormatListQuiet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *ListInProgressResult
		expected string
	}{
		{
			name: "single task",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{
					{ID: 42},
				},
				Count: 1,
			},
			expected: "42\n",
		},
		{
			name: "multiple tasks",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{
					{ID: 1},
					{ID: 2},
					{ID: 3},
				},
				Count: 3,
			},
			expected: "1\n2\n3\n",
		},
		{
			name: "no tasks",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{},
				Count: 0,
			},
			expected: "",
		},
		{
			name: "tasks with large IDs",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{
					{ID: 999999},
					{ID: 1000000},
				},
				Count: 2,
			},
			expected: "999999\n1000000\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatListQuiet(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatListJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *ListInProgressResult
		expected map[string]any
	}{
		{
			name: "single task",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{
					{
						ID:                  42,
						TicketNumber:        1,
						Title:               "Test task",
						TypeDescription:     "feature",
						PriorityDescription: "high",
						PriorityColor:       "#FF0000",
						IsBlocked:           false,
					},
				},
				Count: 1,
			},
			expected: map[string]any{
				"success": true,
				"tasks": []TaskDisplay{
					{
						ID:                  42,
						TicketNumber:        1,
						Title:               "Test task",
						TypeDescription:     "feature",
						PriorityDescription: "high",
						PriorityColor:       "#FF0000",
						IsBlocked:           false,
					},
				},
				"count": 1,
			},
		},
		{
			name: "multiple tasks",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{
					{
						ID:                  1,
						TicketNumber:        1,
						Title:               "Task 1",
						TypeDescription:     "bug",
						PriorityDescription: "critical",
						PriorityColor:       "#FF0000",
						IsBlocked:           true,
					},
					{
						ID:                  2,
						TicketNumber:        2,
						Title:               "Task 2",
						TypeDescription:     "feature",
						PriorityDescription: "medium",
						PriorityColor:       "#00FF00",
						IsBlocked:           false,
					},
				},
				Count: 2,
			},
			expected: map[string]any{
				"success": true,
				"tasks": []TaskDisplay{
					{
						ID:                  1,
						TicketNumber:        1,
						Title:               "Task 1",
						TypeDescription:     "bug",
						PriorityDescription: "critical",
						PriorityColor:       "#FF0000",
						IsBlocked:           true,
					},
					{
						ID:                  2,
						TicketNumber:        2,
						Title:               "Task 2",
						TypeDescription:     "feature",
						PriorityDescription: "medium",
						PriorityColor:       "#00FF00",
						IsBlocked:           false,
					},
				},
				"count": 2,
			},
		},
		{
			name: "no tasks",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{},
				Count: 0,
			},
			expected: map[string]any{
				"success": true,
				"tasks":   []TaskDisplay{},
				"count":   0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatListJSON(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatListHuman(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *ListInProgressResult
		expected string
	}{
		{
			name: "no tasks",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{},
				Count: 0,
			},
			expected: "No in-progress tasks found",
		},
		{
			name: "single task without priority or blocked",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{
					{
						ID:                  42,
						Title:               "Test task",
						PriorityDescription: "",
						IsBlocked:           false,
					},
				},
				Count: 1,
			},
			expected: "Found 1 in-progress tasks:\n\n  [42] Test task\n",
		},
		{
			name: "single task with medium priority (not shown)",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{
					{
						ID:                  42,
						Title:               "Test task",
						PriorityDescription: "medium",
						IsBlocked:           false,
					},
				},
				Count: 1,
			},
			expected: "Found 1 in-progress tasks:\n\n  [42] Test task\n",
		},
		{
			name: "single task with high priority",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{
					{
						ID:                  42,
						Title:               "Test task",
						PriorityDescription: "high",
						IsBlocked:           false,
					},
				},
				Count: 1,
			},
			expected: "Found 1 in-progress tasks:\n\n  [42] Test task [high]\n",
		},
		{
			name: "single task blocked",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{
					{
						ID:                  42,
						Title:               "Test task",
						PriorityDescription: "",
						IsBlocked:           true,
					},
				},
				Count: 1,
			},
			expected: "Found 1 in-progress tasks:\n\n  [42] Test task ▲ BLOCKED\n",
		},
		{
			name: "single task with priority and blocked",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{
					{
						ID:                  42,
						Title:               "Test task",
						PriorityDescription: "critical",
						IsBlocked:           true,
					},
				},
				Count: 1,
			},
			expected: "Found 1 in-progress tasks:\n\n  [42] Test task [critical] ▲ BLOCKED\n",
		},
		{
			name: "multiple tasks with various states",
			result: &ListInProgressResult{
				Tasks: []TaskDisplay{
					{
						ID:                  1,
						Title:               "Task 1",
						PriorityDescription: "high",
						IsBlocked:           false,
					},
					{
						ID:                  2,
						Title:               "Task 2",
						PriorityDescription: "medium",
						IsBlocked:           true,
					},
					{
						ID:                  3,
						Title:               "Task 3",
						PriorityDescription: "",
						IsBlocked:           false,
					},
				},
				Count: 3,
			},
			expected: "Found 3 in-progress tasks:\n\n  [1] Task 1 [high]\n  [2] Task 2 ▲ BLOCKED\n  [3] Task 3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatListHuman(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatMoveQuiet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *MoveInProgressResult
		expected string
	}{
		{
			name: "basic move",
			result: &MoveInProgressResult{
				TaskID:     42,
				FromColumn: "Backlog",
				ToColumn:   "In Progress",
			},
			expected: "42\n",
		},
		{
			name: "move with large task ID",
			result: &MoveInProgressResult{
				TaskID:     999999,
				FromColumn: "To Do",
				ToColumn:   "Doing",
			},
			expected: "999999\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatMoveQuiet(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatMoveJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *MoveInProgressResult
		expected map[string]any
	}{
		{
			name: "basic move",
			result: &MoveInProgressResult{
				TaskID:     42,
				FromColumn: "Backlog",
				ToColumn:   "In Progress",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     42,
				"from_column": "Backlog",
				"to_column":   "In Progress",
			},
		},
		{
			name: "move with different columns",
			result: &MoveInProgressResult{
				TaskID:     100,
				FromColumn: "To Do",
				ToColumn:   "Doing",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     100,
				"from_column": "To Do",
				"to_column":   "Doing",
			},
		},
		{
			name: "move with empty column names",
			result: &MoveInProgressResult{
				TaskID:     1,
				FromColumn: "",
				ToColumn:   "",
			},
			expected: map[string]any{
				"success":     true,
				"task_id":     1,
				"from_column": "",
				"to_column":   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatMoveJSON(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestFormatMoveHuman(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *MoveInProgressResult
		expected string
	}{
		{
			name: "basic move",
			result: &MoveInProgressResult{
				TaskID:     42,
				FromColumn: "Backlog",
				ToColumn:   "In Progress",
			},
			expected: "Task 42 moved to 'In Progress'",
		},
		{
			name: "move with different columns",
			result: &MoveInProgressResult{
				TaskID:     100,
				FromColumn: "To Do",
				ToColumn:   "Doing",
			},
			expected: "Task 100 moved to 'Doing'",
		},
		{
			name: "move with special characters in column name",
			result: &MoveInProgressResult{
				TaskID:     1,
				FromColumn: "Backlog",
				ToColumn:   "In Progress (Sprint 1)",
			},
			expected: "Task 1 moved to 'In Progress (Sprint 1)'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatMoveHuman(tt.result)
			assert.Equal(t, tt.expected, output)
		})
	}
}
