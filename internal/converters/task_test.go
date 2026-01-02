package converters

import (
	"database/sql"
	"testing"
	"time"

	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/models"
)

// ============================================================================
// TEST HELPERS
// ============================================================================

// ptrString returns a pointer to a string
func ptrString(s string) *string {
	return &s
}

// ptrTime returns a pointer to a time.Time
func ptrTime(t time.Time) *time.Time {
	return &t
}

// sep is the separator used in concatenated strings (matching SQL GROUP_CONCAT default)
const sep = ","

// ============================================================================
// TEST CASES - TaskToModel
// ============================================================================

func TestTaskToModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    generated.Task
		expected *models.Task
	}{
		{
			name: "complete task with all fields",
			input: generated.Task{
				ID:          123,
				Title:       "Fix login bug",
				Description: sql.NullString{String: "Users cannot log in", Valid: true},
				ColumnID:    5,
				Position:    10,
				TypeID:      3,
				PriorityID:  4,
				CreatedAt:   sql.NullTime{Time: now, Valid: true},
				UpdatedAt:   sql.NullTime{Time: now, Valid: true},
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
			input: generated.Task{
				ID:          456,
				Title:       "Task without description",
				Description: sql.NullString{Valid: false},
				ColumnID:    2,
				Position:    0,
				TypeID:      1,
				PriorityID:  3,
				CreatedAt:   sql.NullTime{Time: now, Valid: true},
				UpdatedAt:   sql.NullTime{Time: now, Valid: true},
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
			input: generated.Task{
				ID:          789,
				Title:       "Task with null timestamps",
				Description: sql.NullString{String: "Description", Valid: true},
				ColumnID:    1,
				Position:    5,
				TypeID:      2,
				PriorityID:  1,
				CreatedAt:   sql.NullTime{Valid: false},
				UpdatedAt:   sql.NullTime{Valid: false},
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
			input: generated.Task{
				ID:          999,
				Title:       "Minimal task",
				Description: sql.NullString{Valid: false},
				ColumnID:    3,
				Position:    1,
				TypeID:      1,
				PriorityID:  2,
				CreatedAt:   sql.NullTime{Valid: false},
				UpdatedAt:   sql.NullTime{Valid: false},
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
			input: generated.Task{
				ID:          9223372036854775807, // Max int64
				Title:       "Max value task",
				Description: sql.NullString{String: "Testing max values", Valid: true},
				ColumnID:    9223372036854775807,
				Position:    9223372036854775807,
				TypeID:      9223372036854775807,
				PriorityID:  9223372036854775807,
				CreatedAt:   sql.NullTime{Time: now, Valid: true},
				UpdatedAt:   sql.NullTime{Time: now, Valid: true},
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
			input: generated.Task{
				ID:          0,
				Title:       "",
				Description: sql.NullString{String: "", Valid: true},
				ColumnID:    0,
				Position:    0,
				TypeID:      0,
				PriorityID:  0,
				CreatedAt:   sql.NullTime{Time: time.Time{}, Valid: true},
				UpdatedAt:   sql.NullTime{Time: time.Time{}, Valid: true},
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
			input: generated.Task{
				ID:          111,
				Title:       "Task with empty description",
				Description: sql.NullString{String: "", Valid: true},
				ColumnID:    1,
				Position:    0,
				TypeID:      1,
				PriorityID:  1,
				CreatedAt:   sql.NullTime{Time: now, Valid: true},
				UpdatedAt:   sql.NullTime{Time: now, Valid: true},
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

			if result.ID != tt.expected.ID {
				t.Errorf("ID = %d, want %d", result.ID, tt.expected.ID)
			}
			if result.Title != tt.expected.Title {
				t.Errorf("Title = %q, want %q", result.Title, tt.expected.Title)
			}
			if result.Description != tt.expected.Description {
				t.Errorf("Description = %q, want %q", result.Description, tt.expected.Description)
			}
			if result.ColumnID != tt.expected.ColumnID {
				t.Errorf("ColumnID = %d, want %d", result.ColumnID, tt.expected.ColumnID)
			}
			if result.Position != tt.expected.Position {
				t.Errorf("Position = %d, want %d", result.Position, tt.expected.Position)
			}
			if result.TypeID != tt.expected.TypeID {
				t.Errorf("TypeID = %d, want %d", result.TypeID, tt.expected.TypeID)
			}
			if result.PriorityID != tt.expected.PriorityID {
				t.Errorf("PriorityID = %d, want %d", result.PriorityID, tt.expected.PriorityID)
			}
			if !result.CreatedAt.Equal(tt.expected.CreatedAt) {
				t.Errorf("CreatedAt = %v, want %v", result.CreatedAt, tt.expected.CreatedAt)
			}
			if !result.UpdatedAt.Equal(tt.expected.UpdatedAt) {
				t.Errorf("UpdatedAt = %v, want %v", result.UpdatedAt, tt.expected.UpdatedAt)
			}
		})
	}
}

// ============================================================================
// TEST CASES - ParentTasksToReferences
// ============================================================================

func TestParentTasksToReferences(t *testing.T) {
	tests := []struct {
		name     string
		input    []generated.GetParentTasksRow
		expected []*models.TaskReference
	}{
		{
			name:     "empty slice",
			input:    []generated.GetParentTasksRow{},
			expected: []*models.TaskReference{},
		},
		{
			name: "single parent with ticket number",
			input: []generated.GetParentTasksRow{
				{
					ID:           100,
					Title:        "Parent Task",
					Name:         "My Project",
					ID_2:         1,
					PToCLabel:    "Parent of",
					Color:        "#FF0000",
					IsBlocking:   true,
					TicketNumber: sql.NullInt64{Int64: 42, Valid: true},
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
			input: []generated.GetParentTasksRow{
				{
					ID:           200,
					Title:        "Parent Without Ticket",
					Name:         "Another Project",
					ID_2:         2,
					PToCLabel:    "Blocks",
					Color:        "#00FF00",
					IsBlocking:   false,
					TicketNumber: sql.NullInt64{Valid: false},
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
			input: []generated.GetParentTasksRow{
				{
					ID:           1,
					Title:        "First Parent",
					Name:         "Project A",
					ID_2:         1,
					PToCLabel:    "Parent",
					Color:        "#FF0000",
					IsBlocking:   true,
					TicketNumber: sql.NullInt64{Int64: 10, Valid: true},
				},
				{
					ID:           2,
					Title:        "Second Parent",
					Name:         "Project B",
					ID_2:         2,
					PToCLabel:    "Blocker",
					Color:        "#00FF00",
					IsBlocking:   false,
					TicketNumber: sql.NullInt64{Int64: 20, Valid: true},
				},
				{
					ID:           3,
					Title:        "Third Parent",
					Name:         "Project C",
					ID_2:         3,
					PToCLabel:    "Related",
					Color:        "#0000FF",
					IsBlocking:   true,
					TicketNumber: sql.NullInt64{Valid: false},
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

			if len(result) != len(tt.expected) {
				t.Fatalf("length = %d, want %d", len(result), len(tt.expected))
			}

			for i := range result {
				if result[i].ID != tt.expected[i].ID {
					t.Errorf("[%d] ID = %d, want %d", i, result[i].ID, tt.expected[i].ID)
				}
				if result[i].Title != tt.expected[i].Title {
					t.Errorf("[%d] Title = %q, want %q", i, result[i].Title, tt.expected[i].Title)
				}
				if result[i].ProjectName != tt.expected[i].ProjectName {
					t.Errorf("[%d] ProjectName = %q, want %q", i, result[i].ProjectName, tt.expected[i].ProjectName)
				}
				if result[i].RelationTypeID != tt.expected[i].RelationTypeID {
					t.Errorf("[%d] RelationTypeID = %d, want %d", i, result[i].RelationTypeID, tt.expected[i].RelationTypeID)
				}
				if result[i].RelationLabel != tt.expected[i].RelationLabel {
					t.Errorf("[%d] RelationLabel = %q, want %q", i, result[i].RelationLabel, tt.expected[i].RelationLabel)
				}
				if result[i].RelationColor != tt.expected[i].RelationColor {
					t.Errorf("[%d] RelationColor = %q, want %q", i, result[i].RelationColor, tt.expected[i].RelationColor)
				}
				if result[i].IsBlocking != tt.expected[i].IsBlocking {
					t.Errorf("[%d] IsBlocking = %v, want %v", i, result[i].IsBlocking, tt.expected[i].IsBlocking)
				}
				if result[i].TicketNumber != tt.expected[i].TicketNumber {
					t.Errorf("[%d] TicketNumber = %d, want %d", i, result[i].TicketNumber, tt.expected[i].TicketNumber)
				}
			}
		})
	}
}

// ============================================================================
// TEST CASES - ChildTasksToReferences
// ============================================================================

func TestChildTasksToReferences(t *testing.T) {
	tests := []struct {
		name     string
		input    []generated.GetChildTasksRow
		expected []*models.TaskReference
	}{
		{
			name:     "empty slice",
			input:    []generated.GetChildTasksRow{},
			expected: []*models.TaskReference{},
		},
		{
			name: "single child with ticket number",
			input: []generated.GetChildTasksRow{
				{
					ID:           300,
					Title:        "Child Task",
					Name:         "Test Project",
					ID_2:         1,
					CToPLabel:    "Child of",
					Color:        "#FFAA00",
					IsBlocking:   false,
					TicketNumber: sql.NullInt64{Int64: 99, Valid: true},
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
			input: []generated.GetChildTasksRow{
				{
					ID:           400,
					Title:        "Child Without Ticket",
					Name:         "Demo Project",
					ID_2:         3,
					CToPLabel:    "Blocked by",
					Color:        "#AA00FF",
					IsBlocking:   true,
					TicketNumber: sql.NullInt64{Valid: false},
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
			input: []generated.GetChildTasksRow{
				{
					ID:           10,
					Title:        "First Child",
					Name:         "Alpha",
					ID_2:         1,
					CToPLabel:    "Child",
					Color:        "#111111",
					IsBlocking:   false,
					TicketNumber: sql.NullInt64{Int64: 5, Valid: true},
				},
				{
					ID:           20,
					Title:        "Second Child",
					Name:         "Beta",
					ID_2:         2,
					CToPLabel:    "Depends on",
					Color:        "#222222",
					IsBlocking:   true,
					TicketNumber: sql.NullInt64{Int64: 15, Valid: true},
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

			if len(result) != len(tt.expected) {
				t.Fatalf("length = %d, want %d", len(result), len(tt.expected))
			}

			for i := range result {
				if result[i].ID != tt.expected[i].ID {
					t.Errorf("[%d] ID = %d, want %d", i, result[i].ID, tt.expected[i].ID)
				}
				if result[i].Title != tt.expected[i].Title {
					t.Errorf("[%d] Title = %q, want %q", i, result[i].Title, tt.expected[i].Title)
				}
				if result[i].ProjectName != tt.expected[i].ProjectName {
					t.Errorf("[%d] ProjectName = %q, want %q", i, result[i].ProjectName, tt.expected[i].ProjectName)
				}
				if result[i].RelationTypeID != tt.expected[i].RelationTypeID {
					t.Errorf("[%d] RelationTypeID = %d, want %d", i, result[i].RelationTypeID, tt.expected[i].RelationTypeID)
				}
				if result[i].RelationLabel != tt.expected[i].RelationLabel {
					t.Errorf("[%d] RelationLabel = %q, want %q", i, result[i].RelationLabel, tt.expected[i].RelationLabel)
				}
				if result[i].RelationColor != tt.expected[i].RelationColor {
					t.Errorf("[%d] RelationColor = %q, want %q", i, result[i].RelationColor, tt.expected[i].RelationColor)
				}
				if result[i].IsBlocking != tt.expected[i].IsBlocking {
					t.Errorf("[%d] IsBlocking = %v, want %v", i, result[i].IsBlocking, tt.expected[i].IsBlocking)
				}
				if result[i].TicketNumber != tt.expected[i].TicketNumber {
					t.Errorf("[%d] TicketNumber = %d, want %d", i, result[i].TicketNumber, tt.expected[i].TicketNumber)
				}
			}
		})
	}
}

// ============================================================================
// TEST CASES - CommentsToModels
// ============================================================================

func TestCommentsToModels(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    []generated.TaskComment
		expected []*models.Comment
	}{
		{
			name:     "empty slice",
			input:    []generated.TaskComment{},
			expected: []*models.Comment{},
		},
		{
			name: "single comment",
			input: []generated.TaskComment{
				{
					ID:        1,
					TaskID:    100,
					Content:   "This is a comment",
					Author:    "john@example.com",
					CreatedAt: sql.NullTime{Time: now, Valid: true},
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
			input: []generated.TaskComment{
				{
					ID:        1,
					TaskID:    100,
					Content:   "First comment",
					Author:    "alice@example.com",
					CreatedAt: sql.NullTime{Time: now, Valid: true},
				},
				{
					ID:        2,
					TaskID:    100,
					Content:   "Second comment",
					Author:    "bob@example.com",
					CreatedAt: sql.NullTime{Time: now.Add(1 * time.Hour), Valid: true},
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
			input: []generated.TaskComment{
				{
					ID:        99,
					TaskID:    200,
					Content:   "",
					Author:    "test@example.com",
					CreatedAt: sql.NullTime{Time: now, Valid: true},
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

			if len(result) != len(tt.expected) {
				t.Fatalf("length = %d, want %d", len(result), len(tt.expected))
			}

			for i := range result {
				if result[i].ID != tt.expected[i].ID {
					t.Errorf("[%d] ID = %d, want %d", i, result[i].ID, tt.expected[i].ID)
				}
				if result[i].TaskID != tt.expected[i].TaskID {
					t.Errorf("[%d] TaskID = %d, want %d", i, result[i].TaskID, tt.expected[i].TaskID)
				}
				if result[i].Message != tt.expected[i].Message {
					t.Errorf("[%d] Message = %q, want %q", i, result[i].Message, tt.expected[i].Message)
				}
				if result[i].Author != tt.expected[i].Author {
					t.Errorf("[%d] Author = %q, want %q", i, result[i].Author, tt.expected[i].Author)
				}
				if !result[i].CreatedAt.Equal(tt.expected[i].CreatedAt) {
					t.Errorf("[%d] CreatedAt = %v, want %v", i, result[i].CreatedAt, tt.expected[i].CreatedAt)
				}
			}
		})
	}
}

// ============================================================================
// TEST CASES - ParseLabelsFromConcatenated
// ============================================================================

func TestParseLabelsFromConcatenated(t *testing.T) {
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

			if len(result) != len(tt.expected) {
				t.Fatalf("length = %d, want %d", len(result), len(tt.expected))
			}

			for i := range result {
				if result[i].ID != tt.expected[i].ID {
					t.Errorf("[%d] ID = %d, want %d", i, result[i].ID, tt.expected[i].ID)
				}
				if result[i].Name != tt.expected[i].Name {
					t.Errorf("[%d] Name = %q, want %q", i, result[i].Name, tt.expected[i].Name)
				}
				if result[i].Color != tt.expected[i].Color {
					t.Errorf("[%d] Color = %q, want %q", i, result[i].Color, tt.expected[i].Color)
				}
			}
		})
	}
}

// ============================================================================
// TEST CASES - TaskSummaryFromRowToModel
// ============================================================================

func TestTaskSummaryFromRowToModel(t *testing.T) {
	sep := string(rune(31))

	tests := []struct {
		name     string
		input    generated.GetTaskSummariesByProjectRow
		expected *models.TaskSummary
	}{
		{
			name: "complete summary with all fields",
			input: generated.GetTaskSummariesByProjectRow{
				ID:                  100,
				Title:               "Test Task",
				ColumnID:            5,
				Position:            10,
				IsBlocked:           1,
				TypeDescription:     sql.NullString{String: "bug", Valid: true},
				PriorityDescription: sql.NullString{String: "high", Valid: true},
				PriorityColor:       sql.NullString{String: "#FF0000", Valid: true},
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
			input: generated.GetTaskSummariesByProjectRow{
				ID:                  200,
				Title:               "Minimal Task",
				ColumnID:            1,
				Position:            0,
				IsBlocked:           0,
				TypeDescription:     sql.NullString{Valid: false},
				PriorityDescription: sql.NullString{Valid: false},
				PriorityColor:       sql.NullString{Valid: false},
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
			input: generated.GetTaskSummariesByProjectRow{
				ID:                  300,
				Title:               "Blocked Task",
				ColumnID:            2,
				Position:            5,
				IsBlocked:           999, // Any positive value means blocked
				TypeDescription:     sql.NullString{String: "feature", Valid: true},
				PriorityDescription: sql.NullString{String: "low", Valid: true},
				PriorityColor:       sql.NullString{String: "#0000FF", Valid: true},
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
			input: generated.GetTaskSummariesByProjectRow{
				ID:                  400,
				Title:               "No Labels",
				ColumnID:            3,
				Position:            1,
				IsBlocked:           0,
				TypeDescription:     sql.NullString{String: "task", Valid: true},
				PriorityDescription: sql.NullString{String: "medium", Valid: true},
				PriorityColor:       sql.NullString{String: "#FFAA00", Valid: true},
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

			if result.ID != tt.expected.ID {
				t.Errorf("ID = %d, want %d", result.ID, tt.expected.ID)
			}
			if result.Title != tt.expected.Title {
				t.Errorf("Title = %q, want %q", result.Title, tt.expected.Title)
			}
			if result.ColumnID != tt.expected.ColumnID {
				t.Errorf("ColumnID = %d, want %d", result.ColumnID, tt.expected.ColumnID)
			}
			if result.Position != tt.expected.Position {
				t.Errorf("Position = %d, want %d", result.Position, tt.expected.Position)
			}
			if result.IsBlocked != tt.expected.IsBlocked {
				t.Errorf("IsBlocked = %v, want %v", result.IsBlocked, tt.expected.IsBlocked)
			}
			if result.TypeDescription != tt.expected.TypeDescription {
				t.Errorf("TypeDescription = %q, want %q", result.TypeDescription, tt.expected.TypeDescription)
			}
			if result.PriorityDescription != tt.expected.PriorityDescription {
				t.Errorf("PriorityDescription = %q, want %q", result.PriorityDescription, tt.expected.PriorityDescription)
			}
			if result.PriorityColor != tt.expected.PriorityColor {
				t.Errorf("PriorityColor = %q, want %q", result.PriorityColor, tt.expected.PriorityColor)
			}
			if len(result.Labels) != len(tt.expected.Labels) {
				t.Fatalf("Labels length = %d, want %d", len(result.Labels), len(tt.expected.Labels))
			}
			for i := range result.Labels {
				if result.Labels[i].ID != tt.expected.Labels[i].ID {
					t.Errorf("Labels[%d].ID = %d, want %d", i, result.Labels[i].ID, tt.expected.Labels[i].ID)
				}
				if result.Labels[i].Name != tt.expected.Labels[i].Name {
					t.Errorf("Labels[%d].Name = %q, want %q", i, result.Labels[i].Name, tt.expected.Labels[i].Name)
				}
				if result.Labels[i].Color != tt.expected.Labels[i].Color {
					t.Errorf("Labels[%d].Color = %q, want %q", i, result.Labels[i].Color, tt.expected.Labels[i].Color)
				}
			}
		})
	}
}

// ============================================================================
// TEST CASES - ReadyTaskSummaryFromRowToModel
// ============================================================================

func TestReadyTaskSummaryFromRowToModel(t *testing.T) {
	tests := []struct {
		name     string
		input    generated.GetReadyTaskSummariesByProjectRow
		expected *models.TaskSummary
	}{
		{
			name: "ready task with all fields",
			input: generated.GetReadyTaskSummariesByProjectRow{
				ID:                  500,
				Title:               "Ready Task",
				ColumnID:            7,
				Position:            3,
				IsBlocked:           0,
				TypeDescription:     sql.NullString{String: "story", Valid: true},
				PriorityDescription: sql.NullString{String: "critical", Valid: true},
				PriorityColor:       sql.NullString{String: "#AA0000", Valid: true},
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
			input: generated.GetReadyTaskSummariesByProjectRow{
				ID:                  600,
				Title:               "Simple Ready",
				ColumnID:            1,
				Position:            0,
				IsBlocked:           0,
				TypeDescription:     sql.NullString{Valid: false},
				PriorityDescription: sql.NullString{Valid: false},
				PriorityColor:       sql.NullString{Valid: false},
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

			if result.ID != tt.expected.ID {
				t.Errorf("ID = %d, want %d", result.ID, tt.expected.ID)
			}
			if result.Title != tt.expected.Title {
				t.Errorf("Title = %q, want %q", result.Title, tt.expected.Title)
			}
			if result.ColumnID != tt.expected.ColumnID {
				t.Errorf("ColumnID = %d, want %d", result.ColumnID, tt.expected.ColumnID)
			}
			if result.Position != tt.expected.Position {
				t.Errorf("Position = %d, want %d", result.Position, tt.expected.Position)
			}
			if result.IsBlocked != tt.expected.IsBlocked {
				t.Errorf("IsBlocked = %v, want %v", result.IsBlocked, tt.expected.IsBlocked)
			}
			if result.TypeDescription != tt.expected.TypeDescription {
				t.Errorf("TypeDescription = %q, want %q", result.TypeDescription, tt.expected.TypeDescription)
			}
			if result.PriorityDescription != tt.expected.PriorityDescription {
				t.Errorf("PriorityDescription = %q, want %q", result.PriorityDescription, tt.expected.PriorityDescription)
			}
			if result.PriorityColor != tt.expected.PriorityColor {
				t.Errorf("PriorityColor = %q, want %q", result.PriorityColor, tt.expected.PriorityColor)
			}
			if len(result.Labels) != len(tt.expected.Labels) {
				t.Fatalf("Labels length = %d, want %d", len(result.Labels), len(tt.expected.Labels))
			}
		})
	}
}

// ============================================================================
// TEST CASES - FilteredTaskSummaryFromRowToModel
// ============================================================================

func TestFilteredTaskSummaryFromRowToModel(t *testing.T) {
	sep := string(rune(31))

	tests := []struct {
		name     string
		input    generated.GetTaskSummariesByProjectFilteredRow
		expected *models.TaskSummary
	}{
		{
			name: "filtered task with all fields",
			input: generated.GetTaskSummariesByProjectFilteredRow{
				ID:                  700,
				Title:               "Filtered Task",
				ColumnID:            8,
				Position:            2,
				IsBlocked:           1,
				TypeDescription:     sql.NullString{String: "enhancement", Valid: true},
				PriorityDescription: sql.NullString{String: "medium", Valid: true},
				PriorityColor:       sql.NullString{String: "#AAAA00", Valid: true},
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
			input: generated.GetTaskSummariesByProjectFilteredRow{
				ID:                  800,
				Title:               "Simple Filtered",
				ColumnID:            4,
				Position:            1,
				IsBlocked:           0,
				TypeDescription:     sql.NullString{Valid: false},
				PriorityDescription: sql.NullString{Valid: false},
				PriorityColor:       sql.NullString{Valid: false},
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

			if result.ID != tt.expected.ID {
				t.Errorf("ID = %d, want %d", result.ID, tt.expected.ID)
			}
			if result.Title != tt.expected.Title {
				t.Errorf("Title = %q, want %q", result.Title, tt.expected.Title)
			}
			if result.ColumnID != tt.expected.ColumnID {
				t.Errorf("ColumnID = %d, want %d", result.ColumnID, tt.expected.ColumnID)
			}
			if result.Position != tt.expected.Position {
				t.Errorf("Position = %d, want %d", result.Position, tt.expected.Position)
			}
			if result.IsBlocked != tt.expected.IsBlocked {
				t.Errorf("IsBlocked = %v, want %v", result.IsBlocked, tt.expected.IsBlocked)
			}
			if result.TypeDescription != tt.expected.TypeDescription {
				t.Errorf("TypeDescription = %q, want %q", result.TypeDescription, tt.expected.TypeDescription)
			}
			if result.PriorityDescription != tt.expected.PriorityDescription {
				t.Errorf("PriorityDescription = %q, want %q", result.PriorityDescription, tt.expected.PriorityDescription)
			}
			if result.PriorityColor != tt.expected.PriorityColor {
				t.Errorf("PriorityColor = %q, want %q", result.PriorityColor, tt.expected.PriorityColor)
			}
			if len(result.Labels) != len(tt.expected.Labels) {
				t.Fatalf("Labels length = %d, want %d", len(result.Labels), len(tt.expected.Labels))
			}
		})
	}
}
