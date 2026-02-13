package task

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestParseListProjectID(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *ListInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid project ID",
			args: []string{"123"},
			want: &ListInput{ProjectID: 123},
		},
		{
			name: "valid single digit",
			args: []string{"5"},
			want: &ListInput{ProjectID: 5},
		},
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
			errMsg:  "project ID is required",
		},
		{
			name:    "invalid project ID - text",
			args:    []string{"abc"},
			wantErr: true,
			errMsg:  "invalid project ID: abc",
		},
		{
			name:    "invalid project ID - decimal",
			args:    []string{"12.5"},
			wantErr: true,
			errMsg:  "invalid project ID: 12.5",
		},
		{
			name:    "invalid project ID - negative",
			args:    []string{"-1"},
			want:    &ListInput{ProjectID: -1},
			wantErr: false,
		},
		{
			name: "extra arguments ignored",
			args: []string{"42", "extra", "args"},
			want: &ListInput{ProjectID: 42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseListProjectID(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFlattenTasksByColumn(t *testing.T) {
	tests := []struct {
		name          string
		tasksByColumn map[int][]*models.TaskSummary
		expectedCount int
		expectedIDs   []int
	}{
		{
			name: "single column with tasks",
			tasksByColumn: map[int][]*models.TaskSummary{
				1: {
					{ID: 10, Title: "Task 1"},
					{ID: 20, Title: "Task 2"},
				},
			},
			expectedCount: 2,
			expectedIDs:   []int{10, 20},
		},
		{
			name: "multiple columns with tasks",
			tasksByColumn: map[int][]*models.TaskSummary{
				1: {{ID: 10, Title: "Task 1"}},
				2: {{ID: 20, Title: "Task 2"}},
				3: {{ID: 30, Title: "Task 3"}},
			},
			expectedCount: 3,
		},
		{
			name:          "empty map",
			tasksByColumn: map[int][]*models.TaskSummary{},
			expectedCount: 0,
			expectedIDs:   []int{},
		},
		{
			name: "column with empty slice",
			tasksByColumn: map[int][]*models.TaskSummary{
				1: {},
			},
			expectedCount: 0,
		},
		{
			name: "mixed empty and non-empty columns",
			tasksByColumn: map[int][]*models.TaskSummary{
				1: {{ID: 10, Title: "Task 1"}},
				2: {},
				3: {{ID: 30, Title: "Task 3"}},
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenTasksByColumn(tt.tasksByColumn)
			assert.Len(t, result, tt.expectedCount)

			if tt.expectedIDs != nil {
				actualIDs := make([]int, len(result))
				for i, task := range result {
					actualIDs[i] = task.ID
				}
				assert.ElementsMatch(t, tt.expectedIDs, actualIDs)
			}
		})
	}
}

func TestFormatTasksQuiet(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []*models.TaskSummary
		expected []string
	}{
		{
			name: "multiple tasks",
			tasks: []*models.TaskSummary{
				{ID: 1, Title: "Task 1"},
				{ID: 42, Title: "Task 2"},
				{ID: 999, Title: "Task 3"},
			},
			expected: []string{"1", "42", "999"},
		},
		{
			name:     "empty tasks",
			tasks:    []*models.TaskSummary{},
			expected: []string{},
		},
		{
			name: "single task",
			tasks: []*models.TaskSummary{
				{ID: 100, Title: "Solo"},
			},
			expected: []string{"100"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTasksQuiet(tt.tasks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTasksJSON(t *testing.T) {
	tests := []struct {
		name  string
		tasks []*models.TaskSummary
	}{
		{
			name: "with tasks",
			tasks: []*models.TaskSummary{
				{ID: 1, Title: "Task 1"},
				{ID: 2, Title: "Task 2"},
			},
		},
		{
			name:  "empty tasks",
			tasks: []*models.TaskSummary{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTasksJSON(tt.tasks)
			assert.True(t, result["success"].(bool))
			assert.Equal(t, tt.tasks, result["tasks"])
		})
	}
}

func TestBuildTableRows(t *testing.T) {
	estimate1 := "5h"
	estimate2 := ""

	tests := []struct {
		name     string
		tasks    []*models.TaskSummary
		expected []TableRow
	}{
		{
			name: "task with all fields",
			tasks: []*models.TaskSummary{
				{
					ID:                  1,
					Title:               "Fix bug",
					PriorityDescription: "High",
					PriorityColor:       "#FF0000",
					TypeDescription:     "Bug",
					Estimate:            &estimate1,
					Labels: []*models.Label{
						{Name: "backend"},
						{Name: "urgent"},
					},
					IsBlocked: true,
				},
			},
			expected: []TableRow{
				{
					ID:                  "1",
					Title:               "Fix bug",
					PriorityDescription: "High",
					TypeDescription:     "Bug",
					Estimate:            "5h",
					Labels:              "backend, urgent",
					Blocked:             "BLOCKED",
					PriorityColor:       "#FF0000",
				},
			},
		},
		{
			name: "task with no labels",
			tasks: []*models.TaskSummary{
				{
					ID:                  2,
					Title:               "Simple task",
					PriorityDescription: "Low",
					TypeDescription:     "Feature",
					Estimate:            &estimate1,
					Labels:              []*models.Label{},
					IsBlocked:           false,
				},
			},
			expected: []TableRow{
				{
					ID:                  "2",
					Title:               "Simple task",
					PriorityDescription: "Low",
					TypeDescription:     "Feature",
					Estimate:            "5h",
					Labels:              "-",
					Blocked:             "",
					PriorityColor:       "",
				},
			},
		},
		{
			name: "task with nil estimate",
			tasks: []*models.TaskSummary{
				{
					ID:                  3,
					Title:               "No estimate",
					PriorityDescription: "Medium",
					TypeDescription:     "Chore",
					Estimate:            nil,
					Labels:              []*models.Label{},
					IsBlocked:           false,
				},
			},
			expected: []TableRow{
				{
					ID:                  "3",
					Title:               "No estimate",
					PriorityDescription: "Medium",
					TypeDescription:     "Chore",
					Estimate:            "-",
					Labels:              "-",
					Blocked:             "",
					PriorityColor:       "",
				},
			},
		},
		{
			name: "task with empty estimate",
			tasks: []*models.TaskSummary{
				{
					ID:                  4,
					Title:               "Empty estimate",
					PriorityDescription: "Medium",
					TypeDescription:     "Feature",
					Estimate:            &estimate2,
					Labels:              []*models.Label{},
					IsBlocked:           false,
				},
			},
			expected: []TableRow{
				{
					ID:                  "4",
					Title:               "Empty estimate",
					PriorityDescription: "Medium",
					TypeDescription:     "Feature",
					Estimate:            "-",
					Labels:              "-",
					Blocked:             "",
					PriorityColor:       "",
				},
			},
		},
		{
			name: "blocked task",
			tasks: []*models.TaskSummary{
				{
					ID:                  5,
					Title:               "Blocked task",
					PriorityDescription: "High",
					TypeDescription:     "Feature",
					Estimate:            nil,
					Labels:              []*models.Label{},
					IsBlocked:           true,
				},
			},
			expected: []TableRow{
				{
					ID:                  "5",
					Title:               "Blocked task",
					PriorityDescription: "High",
					TypeDescription:     "Feature",
					Estimate:            "-",
					Labels:              "-",
					Blocked:             "BLOCKED",
					PriorityColor:       "",
				},
			},
		},
		{
			name: "task with single label",
			tasks: []*models.TaskSummary{
				{
					ID:                  6,
					Title:               "Single label",
					PriorityDescription: "Low",
					TypeDescription:     "Feature",
					Estimate:            nil,
					Labels: []*models.Label{
						{Name: "frontend"},
					},
					IsBlocked: false,
				},
			},
			expected: []TableRow{
				{
					ID:                  "6",
					Title:               "Single label",
					PriorityDescription: "Low",
					TypeDescription:     "Feature",
					Estimate:            "-",
					Labels:              "frontend",
					Blocked:             "",
					PriorityColor:       "",
				},
			},
		},
		{
			name:     "empty task list",
			tasks:    []*models.TaskSummary{},
			expected: []TableRow{},
		},
		{
			name: "multiple tasks",
			tasks: []*models.TaskSummary{
				{
					ID:                  10,
					Title:               "First",
					PriorityDescription: "High",
					TypeDescription:     "Bug",
					Estimate:            &estimate1,
					Labels:              []*models.Label{{Name: "urgent"}},
					IsBlocked:           false,
				},
				{
					ID:                  20,
					Title:               "Second",
					PriorityDescription: "Low",
					TypeDescription:     "Feature",
					Estimate:            nil,
					Labels:              []*models.Label{},
					IsBlocked:           true,
				},
			},
			expected: []TableRow{
				{
					ID:                  "10",
					Title:               "First",
					PriorityDescription: "High",
					TypeDescription:     "Bug",
					Estimate:            "5h",
					Labels:              "urgent",
					Blocked:             "",
					PriorityColor:       "",
				},
				{
					ID:                  "20",
					Title:               "Second",
					PriorityDescription: "Low",
					TypeDescription:     "Feature",
					Estimate:            "-",
					Labels:              "-",
					Blocked:             "BLOCKED",
					PriorityColor:       "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildTableRows(tt.tasks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderTasksTable(t *testing.T) {
	colorScheme := colors.ColorScheme{
		Normal: "#FFFFFF",
		Accent: "#00FF00",
	}

	tests := []struct {
		name           string
		rows           []TableRow
		colorScheme    colors.ColorScheme
		expectedInText []string
	}{
		{
			name: "basic table rendering",
			rows: []TableRow{
				{
					ID:                  "1",
					Title:               "Test Task",
					PriorityDescription: "High",
					TypeDescription:     "Bug",
					Estimate:            "5h",
					Labels:              "urgent",
					Blocked:             "",
					PriorityColor:       "#FF0000",
				},
			},
			colorScheme: colorScheme,
			expectedInText: []string{
				"Test Task",
				"High",
				"Bug",
				"5h",
				"urgent",
			},
		},
		{
			name: "table with blocked task",
			rows: []TableRow{
				{
					ID:                  "2",
					Title:               "Blocked Task",
					PriorityDescription: "Medium",
					TypeDescription:     "Feature",
					Estimate:            "-",
					Labels:              "-",
					Blocked:             "BLOCKED",
					PriorityColor:       "",
				},
			},
			colorScheme: colorScheme,
			expectedInText: []string{
				"Blocked Task",
				"BLOCKED",
			},
		},
		{
			name: "multiple rows",
			rows: []TableRow{
				{
					ID:                  "1",
					Title:               "First",
					PriorityDescription: "High",
					TypeDescription:     "Bug",
					Estimate:            "3h",
					Labels:              "backend",
					Blocked:             "",
					PriorityColor:       "#FF0000",
				},
				{
					ID:                  "2",
					Title:               "Second",
					PriorityDescription: "Low",
					TypeDescription:     "Feature",
					Estimate:            "-",
					Labels:              "-",
					Blocked:             "BLOCKED",
					PriorityColor:       "",
				},
			},
			colorScheme: colorScheme,
			expectedInText: []string{
				"First",
				"Second",
				"backend",
				"BLOCKED",
			},
		},
		{
			name:        "empty table",
			rows:        []TableRow{},
			colorScheme: colorScheme,
			expectedInText: []string{
				"ID",
				"TITLE",
				"PRIORITY",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderTasksTable(tt.rows, tt.colorScheme)
			assert.NotEmpty(t, result)

			for _, expected := range tt.expectedInText {
				assert.Contains(t, result, expected)
			}

			// Verify headers are present
			assert.Contains(t, result, "ID")
			assert.Contains(t, result, "TITLE")
			assert.Contains(t, result, "PRIORITY")
			assert.Contains(t, result, "TYPE")
			assert.Contains(t, result, "EST")
			assert.Contains(t, result, "LABELS")
		})
	}
}

func TestRenderTasksTableColorScheme(t *testing.T) {
	tests := []struct {
		name        string
		colorScheme colors.ColorScheme
		rows        []TableRow
	}{
		{
			name: "default color scheme",
			colorScheme: colors.ColorScheme{
				Normal: "#FFFFFF",
				Accent: "#00FF00",
			},
			rows: []TableRow{
				{
					ID:                  "1",
					Title:               "Task",
					PriorityDescription: "High",
					TypeDescription:     "Bug",
					Estimate:            "2h",
					Labels:              "test",
					Blocked:             "",
					PriorityColor:       "#FF0000",
				},
			},
		},
		{
			name: "monochrome color scheme",
			colorScheme: colors.ColorScheme{
				Normal: "#888888",
				Accent: "#CCCCCC",
			},
			rows: []TableRow{
				{
					ID:                  "1",
					Title:               "Task",
					PriorityDescription: "Low",
					TypeDescription:     "Feature",
					Estimate:            "-",
					Labels:              "-",
					Blocked:             "",
					PriorityColor:       "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderTasksTable(tt.rows, tt.colorScheme)
			assert.NotEmpty(t, result)
			// Verify the result is a valid table string
			assert.True(t, strings.Contains(result, "ID"))
		})
	}
}

func TestTableRowStructure(t *testing.T) {
	row := TableRow{
		ID:                  "123",
		Title:               "Test",
		PriorityDescription: "High",
		TypeDescription:     "Bug",
		Estimate:            "5h",
		Labels:              "label1, label2",
		Blocked:             "BLOCKED",
		PriorityColor:       "#FF0000",
	}

	assert.Equal(t, "123", row.ID)
	assert.Equal(t, "Test", row.Title)
	assert.Equal(t, "High", row.PriorityDescription)
	assert.Equal(t, "Bug", row.TypeDescription)
	assert.Equal(t, "5h", row.Estimate)
	assert.Equal(t, "label1, label2", row.Labels)
	assert.Equal(t, "BLOCKED", row.Blocked)
	assert.Equal(t, "#FF0000", row.PriorityColor)
}

func TestListInputStructure(t *testing.T) {
	input := ListInput{
		ProjectID: 42,
	}

	assert.Equal(t, 42, input.ProjectID)
}

func TestListResultStructure(t *testing.T) {
	tasks := []*models.TaskSummary{
		{ID: 1, Title: "Task 1"},
		{ID: 2, Title: "Task 2"},
	}

	result := ListResult{
		Tasks: tasks,
	}

	assert.Equal(t, tasks, result.Tasks)
	assert.Len(t, result.Tasks, 2)
}
