package task

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestParseShowTaskID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          []string
		flagID        int
		expected      *ShowInput
		expectedError string
	}{
		{
			name:   "valid positional argument",
			args:   []string{"42"},
			flagID: 0,
			expected: &ShowInput{
				TaskID: 42,
			},
		},
		{
			name:   "valid flag ID when no args",
			args:   []string{},
			flagID: 100,
			expected: &ShowInput{
				TaskID: 100,
			},
		},
		{
			name:   "positional argument takes precedence over flag",
			args:   []string{"42"},
			flagID: 100,
			expected: &ShowInput{
				TaskID: 42,
			},
		},
		{
			name:          "invalid positional argument",
			args:          []string{"not-a-number"},
			flagID:        0,
			expectedError: "task ID must be a positive integer",
		},
		{
			name:          "zero task ID from flag",
			args:          []string{},
			flagID:        0,
			expectedError: "task ID must be a positive integer",
		},
		{
			name:          "negative task ID from flag",
			args:          []string{},
			flagID:        -5,
			expectedError: "task ID must be a positive integer",
		},
		{
			name:          "negative task ID from positional arg",
			args:          []string{"-5"},
			flagID:        0,
			expectedError: "task ID must be a positive integer",
		},
		{
			name:   "large task ID",
			args:   []string{"999999"},
			flagID: 0,
			expected: &ShowInput{
				TaskID: 999999,
			},
		},
		{
			name:          "empty string as positional argument",
			args:          []string{""},
			flagID:        0,
			expectedError: "task ID must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseShowTaskID(tt.args, tt.flagID)

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

func TestFormatShowJSON(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 16, 14, 45, 0, 0, time.UTC)
	commentCreatedAt := time.Date(2024, 1, 17, 9, 0, 0, 0, time.UTC)
	commentUpdatedAt := time.Date(2024, 1, 17, 9, 15, 0, 0, time.UTC)

	assigneeName := "alice"
	estimate := "3h"

	tests := []struct {
		name     string
		task     *models.TaskDetail
		validate func(t *testing.T, result map[string]any)
	}{
		{
			name: "minimal task",
			task: &models.TaskDetail{
				ID:                  1,
				TaskNumber:          42,
				ProjectName:         "TEST",
				Title:               "Test Task",
				Description:         "",
				TypeDescription:     "Feature",
				PriorityDescription: "High",
				PriorityColor:       "#FF0000",
				ColumnID:            10,
				ColumnName:          "In Progress",
				AssigneeName:        nil,
				Estimate:            nil,
				Position:            1,
				IsBlocked:           false,
				Labels:              []*models.Label{},
				ParentTasks:         []*models.TaskReference{},
				ChildTasks:          []*models.TaskReference{},
				Comments:            []*models.Comment{},
				CreatedAt:           createdAt,
				UpdatedAt:           updatedAt,
			},
			validate: func(t *testing.T, result map[string]any) {
				assert.True(t, result["success"].(bool))
				task := result["task"].(map[string]any)
				assert.Equal(t, 1, task["id"])
				assert.Equal(t, 42, task["task_number"])
				assert.Equal(t, "TEST", task["project_name"])
				assert.Equal(t, "Test Task", task["title"])
				assert.Equal(t, "", task["description"])
				assert.Equal(t, "Feature", task["type"])
				assert.Nil(t, task["assignee"])
				assert.Nil(t, task["estimate"])
				assert.False(t, task["is_blocked"].(bool))
				assert.Empty(t, task["labels"])
				assert.Empty(t, task["parent_tasks"])
				assert.Empty(t, task["child_tasks"])
				assert.Empty(t, task["comments"])
			},
		},
		{
			name: "full task with all fields",
			task: &models.TaskDetail{
				ID:                  2,
				TaskNumber:          100,
				ProjectName:         "FULL",
				Title:               "Complete Task",
				Description:         "This is a complete task\nwith multiple lines",
				TypeDescription:     "Bug",
				PriorityDescription: "Critical",
				PriorityColor:       "#FF0000",
				ColumnID:            20,
				ColumnName:          "Done",
				AssigneeName:        &assigneeName,
				Estimate:            &estimate,
				Position:            5,
				IsBlocked:           true,
				Labels: []*models.Label{
					{ID: 1, Name: "frontend", Color: "#0000FF"},
					{ID: 2, Name: "urgent", Color: "#FF0000"},
				},
				ParentTasks: []*models.TaskReference{
					{ID: 50, TaskNumber: 10, ProjectName: "PARENT", Title: "Parent Task", IsBlocking: false, RelationLabel: "subtask"},
				},
				ChildTasks: []*models.TaskReference{
					{ID: 60, TaskNumber: 20, ProjectName: "CHILD", Title: "Child Task", IsBlocking: true, RelationLabel: "blocks"},
				},
				Comments: []*models.Comment{
					{ID: 1, TaskID: 2, Message: "First comment", Author: "bob", CreatedAt: commentCreatedAt, UpdatedAt: commentUpdatedAt},
				},
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			validate: func(t *testing.T, result map[string]any) {
				assert.True(t, result["success"].(bool))
				task := result["task"].(map[string]any)
				assert.Equal(t, 2, task["id"])
				assert.Equal(t, 100, task["task_number"])
				assert.Equal(t, "FULL", task["project_name"])
				assert.Equal(t, "Complete Task", task["title"])
				assert.Equal(t, "This is a complete task\nwith multiple lines", task["description"])
				assert.Equal(t, "Bug", task["type"])
				assert.Equal(t, &assigneeName, task["assignee"])
				assert.Equal(t, &estimate, task["estimate"])
				assert.True(t, task["is_blocked"].(bool))

				priority := task["priority"].(map[string]string)
				assert.Equal(t, "Critical", priority["name"])
				assert.Equal(t, "#FF0000", priority["color"])

				column := task["column"].(map[string]any)
				assert.Equal(t, 20, column["id"])
				assert.Equal(t, "Done", column["name"])

				labels := task["labels"].([]*models.Label)
				assert.Len(t, labels, 2)
				assert.Equal(t, "frontend", labels[0].Name)

				parentTasks := task["parent_tasks"].([]*models.TaskReference)
				assert.Len(t, parentTasks, 1)
				assert.Equal(t, 50, parentTasks[0].ID)

				childTasks := task["child_tasks"].([]*models.TaskReference)
				assert.Len(t, childTasks, 1)
				assert.Equal(t, 60, childTasks[0].ID)

				comments := task["comments"].([]map[string]any)
				assert.Len(t, comments, 1)
				assert.Equal(t, 1, comments[0]["id"])
				assert.Equal(t, "First comment", comments[0]["content"])
				assert.Equal(t, "bob", comments[0]["author"])
			},
		},
		{
			name: "task with multiple comments",
			task: &models.TaskDetail{
				ID:                  3,
				TaskNumber:          200,
				ProjectName:         "COMMENT",
				Title:               "Task with comments",
				TypeDescription:     "Feature",
				PriorityDescription: "Medium",
				PriorityColor:       "#FFFF00",
				ColumnID:            30,
				ColumnName:          "Review",
				Position:            1,
				IsBlocked:           false,
				Labels:              []*models.Label{},
				ParentTasks:         []*models.TaskReference{},
				ChildTasks:          []*models.TaskReference{},
				Comments: []*models.Comment{
					{ID: 1, TaskID: 3, Message: "First", Author: "alice", CreatedAt: commentCreatedAt, UpdatedAt: commentUpdatedAt},
					{ID: 2, TaskID: 3, Message: "Second", Author: "bob", CreatedAt: commentCreatedAt.Add(time.Hour), UpdatedAt: commentUpdatedAt.Add(time.Hour)},
					{ID: 3, TaskID: 3, Message: "Third", Author: "charlie", CreatedAt: commentCreatedAt.Add(2 * time.Hour), UpdatedAt: commentUpdatedAt.Add(2 * time.Hour)},
				},
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			validate: func(t *testing.T, result map[string]any) {
				task := result["task"].(map[string]any)
				comments := task["comments"].([]map[string]any)
				assert.Len(t, comments, 3)
				assert.Equal(t, "First", comments[0]["content"])
				assert.Equal(t, "alice", comments[0]["author"])
				assert.Equal(t, "Second", comments[1]["content"])
				assert.Equal(t, "bob", comments[1]["author"])
				assert.Equal(t, "Third", comments[2]["content"])
				assert.Equal(t, "charlie", comments[2]["author"])
			},
		},
		{
			name: "task with empty comments list",
			task: &models.TaskDetail{
				ID:                  4,
				TaskNumber:          300,
				ProjectName:         "NO-COMMENT",
				Title:               "No comments",
				TypeDescription:     "Feature",
				PriorityDescription: "Low",
				PriorityColor:       "#00FF00",
				ColumnID:            40,
				ColumnName:          "Backlog",
				Position:            1,
				IsBlocked:           false,
				Labels:              []*models.Label{},
				ParentTasks:         []*models.TaskReference{},
				ChildTasks:          []*models.TaskReference{},
				Comments:            []*models.Comment{},
				CreatedAt:           createdAt,
				UpdatedAt:           updatedAt,
			},
			validate: func(t *testing.T, result map[string]any) {
				task := result["task"].(map[string]any)
				comments := task["comments"].([]map[string]any)
				assert.Empty(t, comments)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FormatShowJSON(tt.task)
			tt.validate(t, result)
		})
	}
}

func TestOrganizeTaskRelationships(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		task     *models.TaskDetail
		expected *TaskRelationships
	}{
		{
			name: "no relationships",
			task: &models.TaskDetail{
				ParentTasks: []*models.TaskReference{},
				ChildTasks:  []*models.TaskReference{},
			},
			expected: &TaskRelationships{
				BlockingChildren:    []*models.TaskReference{},
				BlockingParents:     []*models.TaskReference{},
				NonBlockingParents:  []*models.TaskReference{},
				NonBlockingChildren: []*models.TaskReference{},
			},
		},
		{
			name: "only blocking children",
			task: &models.TaskDetail{
				ParentTasks: []*models.TaskReference{},
				ChildTasks: []*models.TaskReference{
					{ID: 1, TaskNumber: 10, ProjectName: "TEST", Title: "Blocker 1", IsBlocking: true, RelationLabel: "blocks"},
					{ID: 2, TaskNumber: 20, ProjectName: "TEST", Title: "Blocker 2", IsBlocking: true, RelationLabel: "blocks"},
				},
			},
			expected: &TaskRelationships{
				BlockingChildren: []*models.TaskReference{
					{ID: 1, TaskNumber: 10, ProjectName: "TEST", Title: "Blocker 1", IsBlocking: true, RelationLabel: "blocks"},
					{ID: 2, TaskNumber: 20, ProjectName: "TEST", Title: "Blocker 2", IsBlocking: true, RelationLabel: "blocks"},
				},
				BlockingParents:     []*models.TaskReference{},
				NonBlockingParents:  []*models.TaskReference{},
				NonBlockingChildren: []*models.TaskReference{},
			},
		},
		{
			name: "only blocking parents",
			task: &models.TaskDetail{
				ParentTasks: []*models.TaskReference{
					{ID: 3, TaskNumber: 30, ProjectName: "TEST", Title: "Blocked 1", IsBlocking: true, RelationLabel: "blocks"},
					{ID: 4, TaskNumber: 40, ProjectName: "TEST", Title: "Blocked 2", IsBlocking: true, RelationLabel: "blocks"},
				},
				ChildTasks: []*models.TaskReference{},
			},
			expected: &TaskRelationships{
				BlockingChildren: []*models.TaskReference{},
				BlockingParents: []*models.TaskReference{
					{ID: 3, TaskNumber: 30, ProjectName: "TEST", Title: "Blocked 1", IsBlocking: true, RelationLabel: "blocks"},
					{ID: 4, TaskNumber: 40, ProjectName: "TEST", Title: "Blocked 2", IsBlocking: true, RelationLabel: "blocks"},
				},
				NonBlockingParents:  []*models.TaskReference{},
				NonBlockingChildren: []*models.TaskReference{},
			},
		},
		{
			name: "only non-blocking parents",
			task: &models.TaskDetail{
				ParentTasks: []*models.TaskReference{
					{ID: 5, TaskNumber: 50, ProjectName: "TEST", Title: "Parent 1", IsBlocking: false, RelationLabel: "subtask"},
					{ID: 6, TaskNumber: 60, ProjectName: "TEST", Title: "Parent 2", IsBlocking: false, RelationLabel: "subtask"},
				},
				ChildTasks: []*models.TaskReference{},
			},
			expected: &TaskRelationships{
				BlockingChildren: []*models.TaskReference{},
				BlockingParents:  []*models.TaskReference{},
				NonBlockingParents: []*models.TaskReference{
					{ID: 5, TaskNumber: 50, ProjectName: "TEST", Title: "Parent 1", IsBlocking: false, RelationLabel: "subtask"},
					{ID: 6, TaskNumber: 60, ProjectName: "TEST", Title: "Parent 2", IsBlocking: false, RelationLabel: "subtask"},
				},
				NonBlockingChildren: []*models.TaskReference{},
			},
		},
		{
			name: "only non-blocking children",
			task: &models.TaskDetail{
				ParentTasks: []*models.TaskReference{},
				ChildTasks: []*models.TaskReference{
					{ID: 7, TaskNumber: 70, ProjectName: "TEST", Title: "Child 1", IsBlocking: false, RelationLabel: "subtask"},
					{ID: 8, TaskNumber: 80, ProjectName: "TEST", Title: "Child 2", IsBlocking: false, RelationLabel: "subtask"},
				},
			},
			expected: &TaskRelationships{
				BlockingChildren:   []*models.TaskReference{},
				BlockingParents:    []*models.TaskReference{},
				NonBlockingParents: []*models.TaskReference{},
				NonBlockingChildren: []*models.TaskReference{
					{ID: 7, TaskNumber: 70, ProjectName: "TEST", Title: "Child 1", IsBlocking: false, RelationLabel: "subtask"},
					{ID: 8, TaskNumber: 80, ProjectName: "TEST", Title: "Child 2", IsBlocking: false, RelationLabel: "subtask"},
				},
			},
		},
		{
			name: "mixed relationships",
			task: &models.TaskDetail{
				ParentTasks: []*models.TaskReference{
					{ID: 9, TaskNumber: 90, ProjectName: "TEST", Title: "Blocking Parent", IsBlocking: true, RelationLabel: "blocks"},
					{ID: 10, TaskNumber: 100, ProjectName: "TEST", Title: "Non-blocking Parent", IsBlocking: false, RelationLabel: "subtask"},
				},
				ChildTasks: []*models.TaskReference{
					{ID: 11, TaskNumber: 110, ProjectName: "TEST", Title: "Blocking Child", IsBlocking: true, RelationLabel: "blocks"},
					{ID: 12, TaskNumber: 120, ProjectName: "TEST", Title: "Non-blocking Child", IsBlocking: false, RelationLabel: "subtask"},
				},
			},
			expected: &TaskRelationships{
				BlockingChildren: []*models.TaskReference{
					{ID: 11, TaskNumber: 110, ProjectName: "TEST", Title: "Blocking Child", IsBlocking: true, RelationLabel: "blocks"},
				},
				BlockingParents: []*models.TaskReference{
					{ID: 9, TaskNumber: 90, ProjectName: "TEST", Title: "Blocking Parent", IsBlocking: true, RelationLabel: "blocks"},
				},
				NonBlockingParents: []*models.TaskReference{
					{ID: 10, TaskNumber: 100, ProjectName: "TEST", Title: "Non-blocking Parent", IsBlocking: false, RelationLabel: "subtask"},
				},
				NonBlockingChildren: []*models.TaskReference{
					{ID: 12, TaskNumber: 120, ProjectName: "TEST", Title: "Non-blocking Child", IsBlocking: false, RelationLabel: "subtask"},
				},
			},
		},
		{
			name: "multiple mixed relationships",
			task: &models.TaskDetail{
				ParentTasks: []*models.TaskReference{
					{ID: 13, TaskNumber: 130, ProjectName: "TEST", Title: "Blocking Parent 1", IsBlocking: true, RelationLabel: "blocks"},
					{ID: 14, TaskNumber: 140, ProjectName: "TEST", Title: "Blocking Parent 2", IsBlocking: true, RelationLabel: "blocks"},
					{ID: 15, TaskNumber: 150, ProjectName: "TEST", Title: "Non-blocking Parent 1", IsBlocking: false, RelationLabel: "subtask"},
					{ID: 16, TaskNumber: 160, ProjectName: "TEST", Title: "Non-blocking Parent 2", IsBlocking: false, RelationLabel: "subtask"},
				},
				ChildTasks: []*models.TaskReference{
					{ID: 17, TaskNumber: 170, ProjectName: "TEST", Title: "Blocking Child 1", IsBlocking: true, RelationLabel: "blocks"},
					{ID: 18, TaskNumber: 180, ProjectName: "TEST", Title: "Blocking Child 2", IsBlocking: true, RelationLabel: "blocks"},
					{ID: 19, TaskNumber: 190, ProjectName: "TEST", Title: "Non-blocking Child 1", IsBlocking: false, RelationLabel: "subtask"},
					{ID: 20, TaskNumber: 200, ProjectName: "TEST", Title: "Non-blocking Child 2", IsBlocking: false, RelationLabel: "subtask"},
				},
			},
			expected: &TaskRelationships{
				BlockingChildren: []*models.TaskReference{
					{ID: 17, TaskNumber: 170, ProjectName: "TEST", Title: "Blocking Child 1", IsBlocking: true, RelationLabel: "blocks"},
					{ID: 18, TaskNumber: 180, ProjectName: "TEST", Title: "Blocking Child 2", IsBlocking: true, RelationLabel: "blocks"},
				},
				BlockingParents: []*models.TaskReference{
					{ID: 13, TaskNumber: 130, ProjectName: "TEST", Title: "Blocking Parent 1", IsBlocking: true, RelationLabel: "blocks"},
					{ID: 14, TaskNumber: 140, ProjectName: "TEST", Title: "Blocking Parent 2", IsBlocking: true, RelationLabel: "blocks"},
				},
				NonBlockingParents: []*models.TaskReference{
					{ID: 15, TaskNumber: 150, ProjectName: "TEST", Title: "Non-blocking Parent 1", IsBlocking: false, RelationLabel: "subtask"},
					{ID: 16, TaskNumber: 160, ProjectName: "TEST", Title: "Non-blocking Parent 2", IsBlocking: false, RelationLabel: "subtask"},
				},
				NonBlockingChildren: []*models.TaskReference{
					{ID: 19, TaskNumber: 190, ProjectName: "TEST", Title: "Non-blocking Child 1", IsBlocking: false, RelationLabel: "subtask"},
					{ID: 20, TaskNumber: 200, ProjectName: "TEST", Title: "Non-blocking Child 2", IsBlocking: false, RelationLabel: "subtask"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := OrganizeTaskRelationships(tt.task)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatShowHuman(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 16, 14, 45, 0, 0, time.UTC)
	commentCreatedAt := time.Date(2024, 1, 17, 9, 0, 0, 0, time.UTC)
	commentUpdatedAt := time.Date(2024, 1, 17, 9, 15, 0, 0, time.UTC)

	assigneeName := "alice"
	estimate := "3h"

	colorScheme := config.DefaultColorScheme()

	tests := []struct {
		name     string
		task     *models.TaskDetail
		validate func(t *testing.T, output string)
	}{
		{
			name: "minimal task",
			task: &models.TaskDetail{
				ID:                  1,
				TaskNumber:          42,
				ProjectName:         "TEST",
				Title:               "Test Task",
				Description:         "",
				TypeDescription:     "Feature",
				PriorityDescription: "High",
				PriorityColor:       "#FF0000",
				ColumnID:            10,
				ColumnName:          "In Progress",
				AssigneeName:        nil,
				Estimate:            nil,
				Position:            1,
				IsBlocked:           false,
				Labels:              []*models.Label{},
				ParentTasks:         []*models.TaskReference{},
				ChildTasks:          []*models.TaskReference{},
				Comments:            []*models.Comment{},
				CreatedAt:           createdAt,
				UpdatedAt:           updatedAt,
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "TEST-42")
				assert.Contains(t, output, "Test Task")
				assert.Contains(t, output, "Feature")
				assert.Contains(t, output, "High")
				assert.Contains(t, output, "In Progress")
				assert.Contains(t, output, "None") // Assignee
				assert.NotContains(t, output, "BLOCKED")
			},
		},
		{
			name: "blocked task",
			task: &models.TaskDetail{
				ID:                  2,
				TaskNumber:          100,
				ProjectName:         "BLOCK",
				Title:               "Blocked Task",
				TypeDescription:     "Bug",
				PriorityDescription: "Critical",
				PriorityColor:       "#FF0000",
				ColumnID:            20,
				ColumnName:          "Blocked",
				Position:            1,
				IsBlocked:           true,
				Labels:              []*models.Label{},
				ParentTasks:         []*models.TaskReference{},
				ChildTasks:          []*models.TaskReference{},
				Comments:            []*models.Comment{},
				CreatedAt:           createdAt,
				UpdatedAt:           updatedAt,
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "BLOCKED")
				assert.Contains(t, output, "BLOCK-100")
			},
		},
		{
			name: "task with description",
			task: &models.TaskDetail{
				ID:                  3,
				TaskNumber:          200,
				ProjectName:         "DESC",
				Title:               "Task with Description",
				Description:         "Line 1\nLine 2\nLine 3",
				TypeDescription:     "Feature",
				PriorityDescription: "Medium",
				PriorityColor:       "#FFFF00",
				ColumnID:            30,
				ColumnName:          "Todo",
				Position:            1,
				IsBlocked:           false,
				Labels:              []*models.Label{},
				ParentTasks:         []*models.TaskReference{},
				ChildTasks:          []*models.TaskReference{},
				Comments:            []*models.Comment{},
				CreatedAt:           createdAt,
				UpdatedAt:           updatedAt,
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Description")
				assert.Contains(t, output, "Line 1")
				assert.Contains(t, output, "Line 2")
				assert.Contains(t, output, "Line 3")
			},
		},
		{
			name: "task with assignee and estimate",
			task: &models.TaskDetail{
				ID:                  4,
				TaskNumber:          300,
				ProjectName:         "ASSIGN",
				Title:               "Assigned Task",
				TypeDescription:     "Feature",
				PriorityDescription: "Low",
				PriorityColor:       "#00FF00",
				ColumnID:            40,
				ColumnName:          "In Progress",
				AssigneeName:        &assigneeName,
				Estimate:            &estimate,
				Position:            1,
				IsBlocked:           false,
				Labels:              []*models.Label{},
				ParentTasks:         []*models.TaskReference{},
				ChildTasks:          []*models.TaskReference{},
				Comments:            []*models.Comment{},
				CreatedAt:           createdAt,
				UpdatedAt:           updatedAt,
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "alice")
				assert.Contains(t, output, "3h")
			},
		},
		{
			name: "task with labels",
			task: &models.TaskDetail{
				ID:                  5,
				TaskNumber:          400,
				ProjectName:         "LABEL",
				Title:               "Labeled Task",
				TypeDescription:     "Feature",
				PriorityDescription: "High",
				PriorityColor:       "#FF0000",
				ColumnID:            50,
				ColumnName:          "Review",
				Position:            1,
				IsBlocked:           false,
				Labels: []*models.Label{
					{ID: 1, Name: "frontend", Color: "#0000FF"},
					{ID: 2, Name: "urgent", Color: "#FF0000"},
				},
				ParentTasks: []*models.TaskReference{},
				ChildTasks:  []*models.TaskReference{},
				Comments:    []*models.Comment{},
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Labels")
				assert.Contains(t, output, "frontend")
				assert.Contains(t, output, "urgent")
			},
		},
		{
			name: "task with blocking children",
			task: &models.TaskDetail{
				ID:                  6,
				TaskNumber:          500,
				ProjectName:         "BLOCKER",
				Title:               "Task blocked by children",
				TypeDescription:     "Feature",
				PriorityDescription: "High",
				PriorityColor:       "#FF0000",
				ColumnID:            60,
				ColumnName:          "Blocked",
				Position:            1,
				IsBlocked:           true,
				Labels:              []*models.Label{},
				ParentTasks:         []*models.TaskReference{},
				ChildTasks: []*models.TaskReference{
					{ID: 1, TaskNumber: 10, ProjectName: "CHILD", Title: "Blocking Child", IsBlocking: true, RelationLabel: "blocks"},
				},
				Comments:  []*models.Comment{},
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Blocked By")
				assert.Contains(t, output, "CHILD-10")
				assert.Contains(t, output, "Blocking Child")
			},
		},
		{
			name: "task with blocking parents",
			task: &models.TaskDetail{
				ID:                  7,
				TaskNumber:          600,
				ProjectName:         "BLOCKER",
				Title:               "Task blocking parents",
				TypeDescription:     "Feature",
				PriorityDescription: "High",
				PriorityColor:       "#FF0000",
				ColumnID:            70,
				ColumnName:          "In Progress",
				Position:            1,
				IsBlocked:           false,
				Labels:              []*models.Label{},
				ParentTasks: []*models.TaskReference{
					{ID: 2, TaskNumber: 20, ProjectName: "PARENT", Title: "Blocked Parent", IsBlocking: true, RelationLabel: "blocks"},
				},
				ChildTasks: []*models.TaskReference{},
				Comments:   []*models.Comment{},
				CreatedAt:  createdAt,
				UpdatedAt:  updatedAt,
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Blocking")
				assert.Contains(t, output, "PARENT-20")
				assert.Contains(t, output, "Blocked Parent")
			},
		},
		{
			name: "task with non-blocking parents",
			task: &models.TaskDetail{
				ID:                  8,
				TaskNumber:          700,
				ProjectName:         "PARENT",
				Title:               "Subtask",
				TypeDescription:     "Feature",
				PriorityDescription: "Medium",
				PriorityColor:       "#FFFF00",
				ColumnID:            80,
				ColumnName:          "In Progress",
				Position:            1,
				IsBlocked:           false,
				Labels:              []*models.Label{},
				ParentTasks: []*models.TaskReference{
					{ID: 3, TaskNumber: 30, ProjectName: "PARENT", Title: "Parent Task", IsBlocking: false, RelationLabel: "subtask"},
				},
				ChildTasks: []*models.TaskReference{},
				Comments:   []*models.Comment{},
				CreatedAt:  createdAt,
				UpdatedAt:  updatedAt,
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Parent Tasks")
				assert.Contains(t, output, "PARENT-30")
			},
		},
		{
			name: "task with non-blocking children",
			task: &models.TaskDetail{
				ID:                  9,
				TaskNumber:          800,
				ProjectName:         "CHILD",
				Title:               "Parent with subtasks",
				TypeDescription:     "Feature",
				PriorityDescription: "Medium",
				PriorityColor:       "#FFFF00",
				ColumnID:            90,
				ColumnName:          "In Progress",
				Position:            1,
				IsBlocked:           false,
				Labels:              []*models.Label{},
				ParentTasks:         []*models.TaskReference{},
				ChildTasks: []*models.TaskReference{
					{ID: 4, TaskNumber: 40, ProjectName: "CHILD", Title: "Child Task", IsBlocking: false, RelationLabel: "subtask"},
				},
				Comments:  []*models.Comment{},
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Child Tasks")
				assert.Contains(t, output, "CHILD-40")
			},
		},
		{
			name: "task with comments",
			task: &models.TaskDetail{
				ID:                  10,
				TaskNumber:          900,
				ProjectName:         "COMMENT",
				Title:               "Task with comments",
				TypeDescription:     "Feature",
				PriorityDescription: "High",
				PriorityColor:       "#FF0000",
				ColumnID:            100,
				ColumnName:          "Review",
				Position:            1,
				IsBlocked:           false,
				Labels:              []*models.Label{},
				ParentTasks:         []*models.TaskReference{},
				ChildTasks:          []*models.TaskReference{},
				Comments: []*models.Comment{
					{ID: 1, TaskID: 10, Message: "First comment", Author: "bob", CreatedAt: commentCreatedAt, UpdatedAt: commentUpdatedAt},
					{ID: 2, TaskID: 10, Message: "Second comment", Author: "alice", CreatedAt: commentCreatedAt.Add(time.Hour), UpdatedAt: commentUpdatedAt.Add(time.Hour)},
				},
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Comments (2)")
				assert.Contains(t, output, "First comment")
				assert.Contains(t, output, "bob")
				assert.Contains(t, output, "Second comment")
				assert.Contains(t, output, "alice")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := FormatShowHuman(tt.task, colorScheme)
			tt.validate(t, output)
		})
	}
}
