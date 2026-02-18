package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestFindNextColumnName(t *testing.T) {
	t.Parallel()

	next2ID := 2
	next3ID := 3

	tests := []struct {
		name            string
		columns         []*models.Column
		currentColumnID int
		expected        string
	}{
		{
			name: "find next column in middle",
			columns: []*models.Column{
				{ID: 1, Name: "Todo", NextID: &next2ID},
				{ID: 2, Name: "In Progress", NextID: &next3ID, PrevID: new(1)},
				{ID: 3, Name: "Done", PrevID: new(2)},
			},
			currentColumnID: 1,
			expected:        "In Progress",
		},
		{
			name: "last column has no next",
			columns: []*models.Column{
				{ID: 1, Name: "Todo", NextID: &next2ID},
				{ID: 2, Name: "Done", PrevID: new(1)},
			},
			currentColumnID: 2,
			expected:        "Unknown",
		},
		{
			name: "single column",
			columns: []*models.Column{
				{ID: 1, Name: "Todo"},
			},
			currentColumnID: 1,
			expected:        "Unknown",
		},
		{
			name: "next column from second to third",
			columns: []*models.Column{
				{ID: 1, Name: "Todo", NextID: &next2ID},
				{ID: 2, Name: "In Progress", NextID: &next3ID, PrevID: new(1)},
				{ID: 3, Name: "Done", PrevID: new(2)},
			},
			currentColumnID: 2,
			expected:        "Done",
		},
		{
			name:            "empty columns list",
			columns:         []*models.Column{},
			currentColumnID: 1,
			expected:        "Unknown",
		},
		{
			name: "column ID not found",
			columns: []*models.Column{
				{ID: 1, Name: "Todo", NextID: &next2ID},
				{ID: 2, Name: "Done", PrevID: new(1)},
			},
			currentColumnID: 999,
			expected:        "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := findNextColumnName(tt.columns, tt.currentColumnID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindPrevColumnName(t *testing.T) {
	t.Parallel()

	next2ID := 2
	next3ID := 3

	tests := []struct {
		name            string
		columns         []*models.Column
		currentColumnID int
		expected        string
	}{
		{
			name: "find previous column in middle",
			columns: []*models.Column{
				{ID: 1, Name: "Todo", NextID: &next2ID},
				{ID: 2, Name: "In Progress", NextID: &next3ID, PrevID: new(1)},
				{ID: 3, Name: "Done", PrevID: new(2)},
			},
			currentColumnID: 2,
			expected:        "Todo",
		},
		{
			name: "first column has no previous",
			columns: []*models.Column{
				{ID: 1, Name: "Todo", NextID: &next2ID},
				{ID: 2, Name: "Done", PrevID: new(1)},
			},
			currentColumnID: 1,
			expected:        "Unknown",
		},
		{
			name: "single column",
			columns: []*models.Column{
				{ID: 1, Name: "Todo"},
			},
			currentColumnID: 1,
			expected:        "Unknown",
		},
		{
			name: "previous column from third to second",
			columns: []*models.Column{
				{ID: 1, Name: "Todo", NextID: &next2ID},
				{ID: 2, Name: "In Progress", NextID: &next3ID, PrevID: new(1)},
				{ID: 3, Name: "Done", PrevID: new(2)},
			},
			currentColumnID: 3,
			expected:        "In Progress",
		},
		{
			name:            "empty columns list",
			columns:         []*models.Column{},
			currentColumnID: 1,
			expected:        "Unknown",
		},
		{
			name: "column ID not found",
			columns: []*models.Column{
				{ID: 1, Name: "Todo", NextID: &next2ID},
				{ID: 2, Name: "Done", PrevID: new(1)},
			},
			currentColumnID: 999,
			expected:        "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := findPrevColumnName(tt.columns, tt.currentColumnID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseMoveTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		target             string
		expectedType       string
		expectedNormalized string
	}{
		{
			name:               "next lowercase",
			target:             "next",
			expectedType:       "next",
			expectedNormalized: "next",
		},
		{
			name:               "next capitalized",
			target:             "Next",
			expectedType:       "next",
			expectedNormalized: "next",
		},
		{
			name:               "next uppercase",
			target:             "NEXT",
			expectedType:       "next",
			expectedNormalized: "next",
		},
		{
			name:               "prev lowercase",
			target:             "prev",
			expectedType:       "prev",
			expectedNormalized: "prev",
		},
		{
			name:               "prev capitalized",
			target:             "Prev",
			expectedType:       "prev",
			expectedNormalized: "prev",
		},
		{
			name:               "prev uppercase",
			target:             "PREV",
			expectedType:       "prev",
			expectedNormalized: "prev",
		},
		{
			name:               "previous lowercase",
			target:             "previous",
			expectedType:       "prev",
			expectedNormalized: "prev",
		},
		{
			name:               "previous capitalized",
			target:             "Previous",
			expectedType:       "prev",
			expectedNormalized: "prev",
		},
		{
			name:               "previous uppercase",
			target:             "PREVIOUS",
			expectedType:       "prev",
			expectedNormalized: "prev",
		},
		{
			name:               "column name lowercase",
			target:             "done",
			expectedType:       "column",
			expectedNormalized: "done",
		},
		{
			name:               "column name with spaces",
			target:             "In Progress",
			expectedType:       "column",
			expectedNormalized: "In Progress",
		},
		{
			name:               "column name capitalized",
			target:             "Done",
			expectedType:       "column",
			expectedNormalized: "Done",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			targetType, normalized := parseMoveTarget(tt.target)
			assert.Equal(t, tt.expectedType, targetType)
			assert.Equal(t, tt.expectedNormalized, normalized)
		})
	}
}

func TestValidateMoveTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		targetType        string
		toColumnName      string
		currentColumnName string
		expectedError     string
	}{
		{
			name:              "valid next move",
			targetType:        "next",
			toColumnName:      "In Progress",
			currentColumnName: "Todo",
			expectedError:     "",
		},
		{
			name:              "invalid next move - already at last column",
			targetType:        "next",
			toColumnName:      "Unknown",
			currentColumnName: "Done",
			expectedError:     "task is already in the last column (Done)",
		},
		{
			name:              "valid prev move",
			targetType:        "prev",
			toColumnName:      "Todo",
			currentColumnName: "In Progress",
			expectedError:     "",
		},
		{
			name:              "invalid prev move - already at first column",
			targetType:        "prev",
			toColumnName:      "Unknown",
			currentColumnName: "Todo",
			expectedError:     "task is already in the first column (Todo)",
		},
		{
			name:              "valid column move",
			targetType:        "column",
			toColumnName:      "Done",
			currentColumnName: "Todo",
			expectedError:     "",
		},
		{
			name:              "column type with Unknown doesn't error",
			targetType:        "column",
			toColumnName:      "Unknown",
			currentColumnName: "Todo",
			expectedError:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := validateMoveTarget(tt.targetType, tt.toColumnName, tt.currentColumnName)
			assert.Equal(t, tt.expectedError, result)
		})
	}
}

func TestFormatMoveMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		taskID     int
		fromColumn string
		toColumn   string
		dryRun     bool
		expected   string
	}{
		{
			name:       "normal move",
			taskID:     1,
			fromColumn: "Todo",
			toColumn:   "In Progress",
			dryRun:     false,
			expected:   "Task 1 moved to 'In Progress'",
		},
		{
			name:       "dry run move",
			taskID:     2,
			fromColumn: "Todo",
			toColumn:   "Done",
			dryRun:     true,
			expected:   "Would move task 2 from 'Todo' to 'Done'",
		},
		{
			name:       "already in target column",
			taskID:     3,
			fromColumn: "Done",
			toColumn:   "Done",
			dryRun:     false,
			expected:   "Task 3 is already in 'Done'",
		},
		{
			name:       "dry run already in target column",
			taskID:     4,
			fromColumn: "Todo",
			toColumn:   "Todo",
			dryRun:     true,
			expected:   "Would keep task 4 in 'Todo' (already there)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := formatMoveMessage(tt.taskID, tt.fromColumn, tt.toColumn, tt.dryRun)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateMoveResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		taskID     int
		fromColumn string
		toColumn   string
		dryRun     bool
		expected   MoveResult
	}{
		{
			name:       "normal move result",
			taskID:     1,
			fromColumn: "Todo",
			toColumn:   "In Progress",
			dryRun:     false,
			expected: MoveResult{
				Success:    true,
				TaskID:     1,
				FromColumn: "Todo",
				ToColumn:   "In Progress",
				DryRun:     false,
			},
		},
		{
			name:       "dry run move result",
			taskID:     2,
			fromColumn: "In Progress",
			toColumn:   "Done",
			dryRun:     true,
			expected: MoveResult{
				Success:    true,
				TaskID:     2,
				FromColumn: "In Progress",
				ToColumn:   "Done",
				DryRun:     true,
			},
		},
		{
			name:       "same column move result",
			taskID:     3,
			fromColumn: "Done",
			toColumn:   "Done",
			dryRun:     false,
			expected: MoveResult{
				Success:    true,
				TaskID:     3,
				FromColumn: "Done",
				ToColumn:   "Done",
				DryRun:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := createMoveResult(tt.taskID, tt.fromColumn, tt.toColumn, tt.dryRun)
			assert.Equal(t, tt.expected, result)
		})
	}
}
