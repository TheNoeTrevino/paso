package converters

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/models"
)

// TEST CASES - TaskToModel

func TestTaskToModel(t *testing.T) {
	t.Parallel()
	now := time.Now()

	tests := []struct {
		name     string
		input    types.Task
		expected *models.Task
	}{
		{
			name: "complete task with all fields",
			input: types.Task{
				ID:          123,
				Title:       "Fix login bug",
				Description: types.NullString{String: "Users cannot log in", Valid: true},
				ColumnID:    5,
				Position:    10,
				TypeID:      3,
				PriorityID:  4,
				CreatedAt:   types.NullTime{Time: now, Valid: true},
				UpdatedAt:   types.NullTime{Time: now, Valid: true},
			},
			expected: &models.Task{
				ID:          123,
				Title:       "Fix login bug",
				Description: "Users cannot log in",
				ColumnID:    5,
				Position:    10,
				TypeID:      3,
				PriorityID:  4,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		{
			name: "task with null description",
			input: types.Task{
				ID:          456,
				Title:       "Task without description",
				Description: types.NullString{Valid: false},
				ColumnID:    2,
				Position:    0,
				TypeID:      1,
				PriorityID:  3,
				CreatedAt:   types.NullTime{Time: now, Valid: true},
				UpdatedAt:   types.NullTime{Time: now, Valid: true},
			},
			expected: &models.Task{
				ID:          456,
				Title:       "Task without description",
				Description: "", // Empty string for invalid NullString
				ColumnID:    2,
				Position:    0,
				TypeID:      1,
				PriorityID:  3,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		{
			name: "task with null timestamps",
			input: types.Task{
				ID:          789,
				Title:       "Task with null timestamps",
				Description: types.NullString{String: "Description", Valid: true},
				ColumnID:    1,
				Position:    5,
				TypeID:      2,
				PriorityID:  1,
				CreatedAt:   types.NullTime{Valid: false},
				UpdatedAt:   types.NullTime{Valid: false},
			},
			expected: &models.Task{
				ID:          789,
				Title:       "Task with null timestamps",
				Description: "Description",
				ColumnID:    1,
				Position:    5,
				TypeID:      2,
				PriorityID:  1,
				CreatedAt:   time.Time{}, // Zero time for invalid NullTime
				UpdatedAt:   time.Time{},
			},
		},
		{
			name: "task with all null optional fields",
			input: types.Task{
				ID:          999,
				Title:       "Minimal task",
				Description: types.NullString{Valid: false},
				ColumnID:    3,
				Position:    1,
				TypeID:      1,
				PriorityID:  2,
				CreatedAt:   types.NullTime{Valid: false},
				UpdatedAt:   types.NullTime{Valid: false},
			},
			expected: &models.Task{
				ID:          999,
				Title:       "Minimal task",
				Description: "",
				ColumnID:    3,
				Position:    1,
				TypeID:      1,
				PriorityID:  2,
				CreatedAt:   time.Time{},
				UpdatedAt:   time.Time{},
			},
		},
		{
			name: "task with max int64 values",
			input: types.Task{
				ID:          9223372036854775807, // Max int64
				Title:       "Max value task",
				Description: types.NullString{String: "Testing max values", Valid: true},
				ColumnID:    9223372036854775807,
				Position:    9223372036854775807,
				TypeID:      9223372036854775807,
				PriorityID:  9223372036854775807,
				CreatedAt:   types.NullTime{Time: now, Valid: true},
				UpdatedAt:   types.NullTime{Time: now, Valid: true},
			},
			expected: &models.Task{
				ID:          9223372036854775807,
				Title:       "Max value task",
				Description: "Testing max values",
				ColumnID:    9223372036854775807,
				Position:    9223372036854775807,
				TypeID:      9223372036854775807,
				PriorityID:  9223372036854775807,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		{
			name: "task with zero values",
			input: types.Task{
				ID:          0,
				Title:       "",
				Description: types.NullString{String: "", Valid: true},
				ColumnID:    0,
				Position:    0,
				TypeID:      0,
				PriorityID:  0,
				CreatedAt:   types.NullTime{Time: time.Time{}, Valid: true},
				UpdatedAt:   types.NullTime{Time: time.Time{}, Valid: true},
			},
			expected: &models.Task{
				ID:          0,
				Title:       "",
				Description: "",
				ColumnID:    0,
				Position:    0,
				TypeID:      0,
				PriorityID:  0,
				CreatedAt:   time.Time{},
				UpdatedAt:   time.Time{},
			},
		},
		{
			name: "task with empty string description (valid but empty)",
			input: types.Task{
				ID:          111,
				Title:       "Task with empty description",
				Description: types.NullString{String: "", Valid: true},
				ColumnID:    1,
				Position:    0,
				TypeID:      1,
				PriorityID:  1,
				CreatedAt:   types.NullTime{Time: now, Valid: true},
				UpdatedAt:   types.NullTime{Time: now, Valid: true},
			},
			expected: &models.Task{
				ID:          111,
				Title:       "Task with empty description",
				Description: "",
				ColumnID:    1,
				Position:    0,
				TypeID:      1,
				PriorityID:  1,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TaskToModel(tt.input)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Title, result.Title)
			assert.Equal(t, tt.expected.Description, result.Description)
			assert.Equal(t, tt.expected.ColumnID, result.ColumnID)
			assert.Equal(t, tt.expected.Position, result.Position)
			assert.Equal(t, tt.expected.TypeID, result.TypeID)
			assert.Equal(t, tt.expected.PriorityID, result.PriorityID)
			assert.True(t, result.CreatedAt.Equal(tt.expected.CreatedAt))
			assert.True(t, result.UpdatedAt.Equal(tt.expected.UpdatedAt))
		})
	}
}

// TEST CASES - ParentTasksToReferences

func TestParentTasksToReferences(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []types.GetParentTasksRow
		expected []*models.TaskReference
	}{
		{
			name:     "empty slice",
			input:    []types.GetParentTasksRow{},
			expected: []*models.TaskReference{},
		},
		{
			name: "single parent with ticket number",
			input: []types.GetParentTasksRow{
				{
					ID:           100,
					Title:        "Parent Task",
					Name:         "My Project",
					ID_2:         1,
					PToCLabel:    "Parent of",
					Color:        "#FF0000",
					IsBlocking:   true,
					TicketNumber: types.NullInt64{Int64: 42, Valid: true},
				},
			},
			expected: []*models.TaskReference{
				{
					ID:             100,
					Title:          "Parent Task",
					ProjectName:    "My Project",
					RelationTypeID: 1,
					RelationLabel:  "Parent of",
					RelationColor:  "#FF0000",
					IsBlocking:     true,
					TicketNumber:   42,
				},
			},
		},
		{
			name: "parent without ticket number",
			input: []types.GetParentTasksRow{
				{
					ID:           200,
					Title:        "Parent Without Ticket",
					Name:         "Another Project",
					ID_2:         2,
					PToCLabel:    "Blocks",
					Color:        "#00FF00",
					IsBlocking:   false,
					TicketNumber: types.NullInt64{Valid: false},
				},
			},
			expected: []*models.TaskReference{
				{
					ID:             200,
					Title:          "Parent Without Ticket",
					ProjectName:    "Another Project",
					RelationTypeID: 2,
					RelationLabel:  "Blocks",
					RelationColor:  "#00FF00",
					IsBlocking:     false,
					TicketNumber:   0, // Zero value when null
				},
			},
		},
		{
			name: "multiple parents",
			input: []types.GetParentTasksRow{
				{
					ID:           1,
					Title:        "First Parent",
					Name:         "Project A",
					ID_2:         1,
					PToCLabel:    "Parent",
					Color:        "#FF0000",
					IsBlocking:   true,
					TicketNumber: types.NullInt64{Int64: 10, Valid: true},
				},
				{
					ID:           2,
					Title:        "Second Parent",
					Name:         "Project B",
					ID_2:         2,
					PToCLabel:    "Blocker",
					Color:        "#00FF00",
					IsBlocking:   false,
					TicketNumber: types.NullInt64{Int64: 20, Valid: true},
				},
				{
					ID:           3,
					Title:        "Third Parent",
					Name:         "Project C",
					ID_2:         3,
					PToCLabel:    "Related",
					Color:        "#0000FF",
					IsBlocking:   true,
					TicketNumber: types.NullInt64{Valid: false},
				},
			},
			expected: []*models.TaskReference{
				{
					ID:             1,
					Title:          "First Parent",
					ProjectName:    "Project A",
					RelationTypeID: 1,
					RelationLabel:  "Parent",
					RelationColor:  "#FF0000",
					IsBlocking:     true,
					TicketNumber:   10,
				},
				{
					ID:             2,
					Title:          "Second Parent",
					ProjectName:    "Project B",
					RelationTypeID: 2,
					RelationLabel:  "Blocker",
					RelationColor:  "#00FF00",
					IsBlocking:     false,
					TicketNumber:   20,
				},
				{
					ID:             3,
					Title:          "Third Parent",
					ProjectName:    "Project C",
					RelationTypeID: 3,
					RelationLabel:  "Related",
					RelationColor:  "#0000FF",
					IsBlocking:     true,
					TicketNumber:   0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParentTasksToReferences(tt.input)

			require.Len(t, result, len(tt.expected))

			for i := range result {
				assert.Equal(t, tt.expected[i].ID, result[i].ID)
				assert.Equal(t, tt.expected[i].Title, result[i].Title)
				assert.Equal(t, tt.expected[i].ProjectName, result[i].ProjectName)
				assert.Equal(t, tt.expected[i].RelationTypeID, result[i].RelationTypeID)
				assert.Equal(t, tt.expected[i].RelationLabel, result[i].RelationLabel)
				assert.Equal(t, tt.expected[i].RelationColor, result[i].RelationColor)
				assert.Equal(t, tt.expected[i].IsBlocking, result[i].IsBlocking)
				assert.Equal(t, tt.expected[i].TicketNumber, result[i].TicketNumber)
			}
		})
	}
}

// TEST CASES - ChildTasksToReferences

func TestChildTasksToReferences(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []types.GetChildTasksRow
		expected []*models.TaskReference
	}{
		{
			name:     "empty slice",
			input:    []types.GetChildTasksRow{},
			expected: []*models.TaskReference{},
		},
		{
			name: "single child with ticket number",
			input: []types.GetChildTasksRow{
				{
					ID:           300,
					Title:        "Child Task",
					Name:         "Test Project",
					ID_2:         1,
					CToPLabel:    "Child of",
					Color:        "#FFAA00",
					IsBlocking:   false,
					TicketNumber: types.NullInt64{Int64: 99, Valid: true},
				},
			},
			expected: []*models.TaskReference{
				{
					ID:             300,
					Title:          "Child Task",
					ProjectName:    "Test Project",
					RelationTypeID: 1,
					RelationLabel:  "Child of",
					RelationColor:  "#FFAA00",
					IsBlocking:     false,
					TicketNumber:   99,
				},
			},
		},
		{
			name: "child without ticket number",
			input: []types.GetChildTasksRow{
				{
					ID:           400,
					Title:        "Child Without Ticket",
					Name:         "Demo Project",
					ID_2:         3,
					CToPLabel:    "Blocked by",
					Color:        "#AA00FF",
					IsBlocking:   true,
					TicketNumber: types.NullInt64{Valid: false},
				},
			},
			expected: []*models.TaskReference{
				{
					ID:             400,
					Title:          "Child Without Ticket",
					ProjectName:    "Demo Project",
					RelationTypeID: 3,
					RelationLabel:  "Blocked by",
					RelationColor:  "#AA00FF",
					IsBlocking:     true,
					TicketNumber:   0,
				},
			},
		},
		{
			name: "multiple children",
			input: []types.GetChildTasksRow{
				{
					ID:           10,
					Title:        "First Child",
					Name:         "Alpha",
					ID_2:         1,
					CToPLabel:    "Child",
					Color:        "#111111",
					IsBlocking:   false,
					TicketNumber: types.NullInt64{Int64: 5, Valid: true},
				},
				{
					ID:           20,
					Title:        "Second Child",
					Name:         "Beta",
					ID_2:         2,
					CToPLabel:    "Depends on",
					Color:        "#222222",
					IsBlocking:   true,
					TicketNumber: types.NullInt64{Int64: 15, Valid: true},
				},
			},
			expected: []*models.TaskReference{
				{
					ID:             10,
					Title:          "First Child",
					ProjectName:    "Alpha",
					RelationTypeID: 1,
					RelationLabel:  "Child",
					RelationColor:  "#111111",
					IsBlocking:     false,
					TicketNumber:   5,
				},
				{
					ID:             20,
					Title:          "Second Child",
					ProjectName:    "Beta",
					RelationTypeID: 2,
					RelationLabel:  "Depends on",
					RelationColor:  "#222222",
					IsBlocking:     true,
					TicketNumber:   15,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ChildTasksToReferences(tt.input)

			require.Len(t, result, len(tt.expected))

			for i := range result {
				assert.Equal(t, tt.expected[i].ID, result[i].ID)
				assert.Equal(t, tt.expected[i].Title, result[i].Title)
				assert.Equal(t, tt.expected[i].ProjectName, result[i].ProjectName)
				assert.Equal(t, tt.expected[i].RelationTypeID, result[i].RelationTypeID)
				assert.Equal(t, tt.expected[i].RelationLabel, result[i].RelationLabel)
				assert.Equal(t, tt.expected[i].RelationColor, result[i].RelationColor)
				assert.Equal(t, tt.expected[i].IsBlocking, result[i].IsBlocking)
				assert.Equal(t, tt.expected[i].TicketNumber, result[i].TicketNumber)
			}
		})
	}
}

// TEST CASES - CommentsToModels

func TestCommentsToModels(t *testing.T) {
	t.Parallel()
	now := time.Now()

	tests := []struct {
		name     string
		input    []types.TaskComment
		expected []*models.Comment
	}{
		{
			name:     "empty slice",
			input:    []types.TaskComment{},
			expected: []*models.Comment{},
		},
		{
			name: "single comment",
			input: []types.TaskComment{
				{
					ID:        1,
					TaskID:    100,
					Content:   "This is a comment",
					Author:    "john@example.com",
					CreatedAt: types.NullTime{Time: now, Valid: true},
				},
			},
			expected: []*models.Comment{
				{
					ID:        1,
					TaskID:    100,
					Message:   "This is a comment",
					Author:    "john@example.com",
					CreatedAt: now,
				},
			},
		},
		{
			name: "multiple comments",
			input: []types.TaskComment{
				{
					ID:        1,
					TaskID:    100,
					Content:   "First comment",
					Author:    "alice@example.com",
					CreatedAt: types.NullTime{Time: now, Valid: true},
				},
				{
					ID:        2,
					TaskID:    100,
					Content:   "Second comment",
					Author:    "bob@example.com",
					CreatedAt: types.NullTime{Time: now.Add(1 * time.Hour), Valid: true},
				},
			},
			expected: []*models.Comment{
				{
					ID:        1,
					TaskID:    100,
					Message:   "First comment",
					Author:    "alice@example.com",
					CreatedAt: now,
				},
				{
					ID:        2,
					TaskID:    100,
					Message:   "Second comment",
					Author:    "bob@example.com",
					CreatedAt: now.Add(1 * time.Hour),
				},
			},
		},
		{
			name: "comment with empty content",
			input: []types.TaskComment{
				{
					ID:        99,
					TaskID:    200,
					Content:   "",
					Author:    "test@example.com",
					CreatedAt: types.NullTime{Time: now, Valid: true},
				},
			},
			expected: []*models.Comment{
				{
					ID:        99,
					TaskID:    200,
					Message:   "",
					Author:    "test@example.com",
					CreatedAt: now,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CommentsToModels(tt.input)

			require.Len(t, result, len(tt.expected))

			for i := range result {
				assert.Equal(t, tt.expected[i].ID, result[i].ID)
				assert.Equal(t, tt.expected[i].TaskID, result[i].TaskID)
				assert.Equal(t, tt.expected[i].Message, result[i].Message)
				assert.Equal(t, tt.expected[i].Author, result[i].Author)
				assert.True(t, result[i].CreatedAt.Equal(tt.expected[i].CreatedAt))
			}
		})
	}
}

// TEST CASES - ParseLabelsFromConcatenated

func TestParseLabelsFromConcatenated(t *testing.T) {
	t.Parallel()
	sep := string(rune(31)) // labelSeparator

	tests := []struct {
		name     string
		ids      string
		names    string
		colors   string
		expected []*models.Label
	}{
		{
			name:     "empty strings",
			ids:      "",
			names:    "",
			colors:   "",
			expected: []*models.Label{},
		},
		{
			name:     "empty ids only",
			ids:      "",
			names:    "bug",
			colors:   "#FF0000",
			expected: []*models.Label{},
		},
		{
			name:     "empty names only",
			ids:      "1",
			names:    "",
			colors:   "#FF0000",
			expected: []*models.Label{},
		},
		{
			name:     "empty colors only",
			ids:      "1",
			names:    "bug",
			colors:   "",
			expected: []*models.Label{},
		},
		{
			name:   "single label",
			ids:    "1",
			names:  "bug",
			colors: "#FF0000",
			expected: []*models.Label{
				{ID: 1, Name: "bug", Color: "#FF0000"},
			},
		},
		{
			name:   "multiple labels",
			ids:    "1" + sep + "2" + sep + "3",
			names:  "bug" + sep + "feature" + sep + "critical",
			colors: "#FF0000" + sep + "#00FF00" + sep + "#0000FF",
			expected: []*models.Label{
				{ID: 1, Name: "bug", Color: "#FF0000"},
				{ID: 2, Name: "feature", Color: "#00FF00"},
				{ID: 3, Name: "critical", Color: "#0000FF"},
			},
		},
		{
			name:     "mismatched lengths - ids longer",
			ids:      "1" + sep + "2" + sep + "3",
			names:    "bug" + sep + "feature",
			colors:   "#FF0000" + sep + "#00FF00",
			expected: []*models.Label{},
		},
		{
			name:     "mismatched lengths - names longer",
			ids:      "1" + sep + "2",
			names:    "bug" + sep + "feature" + sep + "critical",
			colors:   "#FF0000" + sep + "#00FF00",
			expected: []*models.Label{},
		},
		{
			name:     "mismatched lengths - colors longer",
			ids:      "1" + sep + "2",
			names:    "bug" + sep + "feature",
			colors:   "#FF0000" + sep + "#00FF00" + sep + "#0000FF",
			expected: []*models.Label{},
		},
		{
			name:   "label with non-numeric ID (parses as 0)",
			ids:    "abc",
			names:  "bug",
			colors: "#FF0000",
			expected: []*models.Label{
				{ID: 0, Name: "bug", Color: "#FF0000"},
			},
		},
		{
			name:   "labels with empty name and color",
			ids:    "1" + sep + "2",
			names:  "" + sep + "",
			colors: "" + sep + "",
			expected: []*models.Label{
				{ID: 1, Name: "", Color: ""},
				{ID: 2, Name: "", Color: ""},
			},
		},
		{
			name:   "labels with large IDs",
			ids:    "999999" + sep + "123456789",
			names:  "label1" + sep + "label2",
			colors: "#AAAAAA" + sep + "#BBBBBB",
			expected: []*models.Label{
				{ID: 999999, Name: "label1", Color: "#AAAAAA"},
				{ID: 123456789, Name: "label2", Color: "#BBBBBB"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLabelsFromConcatenated(tt.ids, tt.names, tt.colors)

			require.Len(t, result, len(tt.expected))

			for i := range result {
				assert.Equal(t, tt.expected[i].ID, result[i].ID)
				assert.Equal(t, tt.expected[i].Name, result[i].Name)
				assert.Equal(t, tt.expected[i].Color, result[i].Color)
			}
		})
	}
}

// TEST CASES - TaskSummaryFromRowToModel

func TestTaskSummaryFromRowToModel(t *testing.T) {
	t.Parallel()
	sep := string(rune(31))

	tests := []struct {
		name     string
		input    types.GetTaskSummariesByProjectRow
		expected *models.TaskSummary
	}{
		{
			name: "complete summary with all fields",
			input: types.GetTaskSummariesByProjectRow{
				ID:                  100,
				Title:               "Test Task",
				ColumnID:            5,
				Position:            10,
				IsBlocked:           true,
				TypeDescription:     types.NullString{String: "bug", Valid: true},
				PriorityDescription: types.NullString{String: "high", Valid: true},
				PriorityColor:       types.NullString{String: "#FF0000", Valid: true},
				LabelIds:            "1" + sep + "2",
				LabelNames:          "urgent" + sep + "backend",
				LabelColors:         "#FF0000" + sep + "#00FF00",
			},
			expected: &models.TaskSummary{
				ID:                  100,
				Title:               "Test Task",
				ColumnID:            5,
				Position:            10,
				IsBlocked:           true,
				TypeDescription:     "bug",
				PriorityDescription: "high",
				PriorityColor:       "#FF0000",
				Labels: []*models.Label{
					{ID: 1, Name: "urgent", Color: "#FF0000"},
					{ID: 2, Name: "backend", Color: "#00FF00"},
				},
			},
		},
		{
			name: "summary with null optional fields",
			input: types.GetTaskSummariesByProjectRow{
				ID:                  200,
				Title:               "Minimal Task",
				ColumnID:            1,
				Position:            0,
				IsBlocked:           false,
				TypeDescription:     types.NullString{Valid: false},
				PriorityDescription: types.NullString{Valid: false},
				PriorityColor:       types.NullString{Valid: false},
				LabelIds:            "",
				LabelNames:          "",
				LabelColors:         "",
			},
			expected: &models.TaskSummary{
				ID:                  200,
				Title:               "Minimal Task",
				ColumnID:            1,
				Position:            0,
				IsBlocked:           false,
				TypeDescription:     "",
				PriorityDescription: "",
				PriorityColor:       "",
				Labels:              []*models.Label{},
			},
		},
		{
			name: "summary with blocked task",
			input: types.GetTaskSummariesByProjectRow{
				ID:                  300,
				Title:               "Blocked Task",
				ColumnID:            2,
				Position:            5,
				IsBlocked:           true, // Any positive value means blocked
				TypeDescription:     types.NullString{String: "feature", Valid: true},
				PriorityDescription: types.NullString{String: "low", Valid: true},
				PriorityColor:       types.NullString{String: "#0000FF", Valid: true},
				LabelIds:            "10",
				LabelNames:          "blocked",
				LabelColors:         "#AAAAAA",
			},
			expected: &models.TaskSummary{
				ID:                  300,
				Title:               "Blocked Task",
				ColumnID:            2,
				Position:            5,
				IsBlocked:           true,
				TypeDescription:     "feature",
				PriorityDescription: "low",
				PriorityColor:       "#0000FF",
				Labels: []*models.Label{
					{ID: 10, Name: "blocked", Color: "#AAAAAA"},
				},
			},
		},
		{
			name: "summary with no labels",
			input: types.GetTaskSummariesByProjectRow{
				ID:                  400,
				Title:               "No Labels",
				ColumnID:            3,
				Position:            1,
				IsBlocked:           false,
				TypeDescription:     types.NullString{String: "task", Valid: true},
				PriorityDescription: types.NullString{String: "medium", Valid: true},
				PriorityColor:       types.NullString{String: "#FFAA00", Valid: true},
				LabelIds:            "",
				LabelNames:          "",
				LabelColors:         "",
			},
			expected: &models.TaskSummary{
				ID:                  400,
				Title:               "No Labels",
				ColumnID:            3,
				Position:            1,
				IsBlocked:           false,
				TypeDescription:     "task",
				PriorityDescription: "medium",
				PriorityColor:       "#FFAA00",
				Labels:              []*models.Label{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TaskSummaryFromRowToModel(tt.input)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Title, result.Title)
			assert.Equal(t, tt.expected.ColumnID, result.ColumnID)
			assert.Equal(t, tt.expected.Position, result.Position)
			assert.Equal(t, tt.expected.IsBlocked, result.IsBlocked)
			assert.Equal(t, tt.expected.TypeDescription, result.TypeDescription)
			assert.Equal(t, tt.expected.PriorityDescription, result.PriorityDescription)
			assert.Equal(t, tt.expected.PriorityColor, result.PriorityColor)
			require.Len(t, result.Labels, len(tt.expected.Labels))
			for i := range result.Labels {
				assert.Equal(t, tt.expected.Labels[i].ID, result.Labels[i].ID)
				assert.Equal(t, tt.expected.Labels[i].Name, result.Labels[i].Name)
				assert.Equal(t, tt.expected.Labels[i].Color, result.Labels[i].Color)
			}
		})
	}
}

// TEST CASES - ReadyTaskSummaryFromRowToModel

func TestReadyTaskSummaryFromRowToModel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    types.GetReadyTaskSummariesByProjectRow
		expected *models.TaskSummary
	}{
		{
			name: "ready task with all fields",
			input: types.GetReadyTaskSummariesByProjectRow{
				ID:                  500,
				Title:               "Ready Task",
				ColumnID:            7,
				Position:            3,
				IsBlocked:           false,
				TypeDescription:     types.NullString{String: "story", Valid: true},
				PriorityDescription: types.NullString{String: "critical", Valid: true},
				PriorityColor:       types.NullString{String: "#AA0000", Valid: true},
				LabelIds:            "5",
				LabelNames:          "ready",
				LabelColors:         "#00AA00",
			},
			expected: &models.TaskSummary{
				ID:                  500,
				Title:               "Ready Task",
				ColumnID:            7,
				Position:            3,
				IsBlocked:           false,
				TypeDescription:     "story",
				PriorityDescription: "critical",
				PriorityColor:       "#AA0000",
				Labels: []*models.Label{
					{ID: 5, Name: "ready", Color: "#00AA00"},
				},
			},
		},
		{
			name: "ready task with null fields",
			input: types.GetReadyTaskSummariesByProjectRow{
				ID:                  600,
				Title:               "Simple Ready",
				ColumnID:            1,
				Position:            0,
				IsBlocked:           false,
				TypeDescription:     types.NullString{Valid: false},
				PriorityDescription: types.NullString{Valid: false},
				PriorityColor:       types.NullString{Valid: false},
				LabelIds:            "",
				LabelNames:          "",
				LabelColors:         "",
			},
			expected: &models.TaskSummary{
				ID:                  600,
				Title:               "Simple Ready",
				ColumnID:            1,
				Position:            0,
				IsBlocked:           false,
				TypeDescription:     "",
				PriorityDescription: "",
				PriorityColor:       "",
				Labels:              []*models.Label{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReadyTaskSummaryFromRowToModel(tt.input)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Title, result.Title)
			assert.Equal(t, tt.expected.ColumnID, result.ColumnID)
			assert.Equal(t, tt.expected.Position, result.Position)
			assert.Equal(t, tt.expected.IsBlocked, result.IsBlocked)
			assert.Equal(t, tt.expected.TypeDescription, result.TypeDescription)
			assert.Equal(t, tt.expected.PriorityDescription, result.PriorityDescription)
			assert.Equal(t, tt.expected.PriorityColor, result.PriorityColor)
			require.Len(t, result.Labels, len(tt.expected.Labels))
		})
	}
}

// TEST CASES - FilteredTaskSummaryFromRowToModel

func TestFilteredTaskSummaryFromRowToModel(t *testing.T) {
	t.Parallel()
	sep := string(rune(31))

	tests := []struct {
		name     string
		input    types.GetTaskSummariesByProjectFilteredRow
		expected *models.TaskSummary
	}{
		{
			name: "filtered task with all fields",
			input: types.GetTaskSummariesByProjectFilteredRow{
				ID:                  700,
				Title:               "Filtered Task",
				ColumnID:            8,
				Position:            2,
				IsBlocked:           true,
				TypeDescription:     types.NullString{String: "enhancement", Valid: true},
				PriorityDescription: types.NullString{String: "medium", Valid: true},
				PriorityColor:       types.NullString{String: "#AAAA00", Valid: true},
				LabelIds:            "7" + sep + "8",
				LabelNames:          "filtered" + sep + "test",
				LabelColors:         "#123456" + sep + "#654321",
			},
			expected: &models.TaskSummary{
				ID:                  700,
				Title:               "Filtered Task",
				ColumnID:            8,
				Position:            2,
				IsBlocked:           true,
				TypeDescription:     "enhancement",
				PriorityDescription: "medium",
				PriorityColor:       "#AAAA00",
				Labels: []*models.Label{
					{ID: 7, Name: "filtered", Color: "#123456"},
					{ID: 8, Name: "test", Color: "#654321"},
				},
			},
		},
		{
			name: "filtered task with null fields",
			input: types.GetTaskSummariesByProjectFilteredRow{
				ID:                  800,
				Title:               "Simple Filtered",
				ColumnID:            4,
				Position:            1,
				IsBlocked:           false,
				TypeDescription:     types.NullString{Valid: false},
				PriorityDescription: types.NullString{Valid: false},
				PriorityColor:       types.NullString{Valid: false},
				LabelIds:            "",
				LabelNames:          "",
				LabelColors:         "",
			},
			expected: &models.TaskSummary{
				ID:                  800,
				Title:               "Simple Filtered",
				ColumnID:            4,
				Position:            1,
				IsBlocked:           false,
				TypeDescription:     "",
				PriorityDescription: "",
				PriorityColor:       "",
				Labels:              []*models.Label{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilteredTaskSummaryFromRowToModel(tt.input)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Title, result.Title)
			assert.Equal(t, tt.expected.ColumnID, result.ColumnID)
			assert.Equal(t, tt.expected.Position, result.Position)
			assert.Equal(t, tt.expected.IsBlocked, result.IsBlocked)
			assert.Equal(t, tt.expected.TypeDescription, result.TypeDescription)
			assert.Equal(t, tt.expected.PriorityDescription, result.PriorityDescription)
			assert.Equal(t, tt.expected.PriorityColor, result.PriorityColor)
			require.Len(t, result.Labels, len(tt.expected.Labels))
		})
	}
}
