package converters

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestColumnToModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    types.Column
		expected *models.Column
	}{
		{
			name: "basic column with all flags false",
			input: types.Column{
				ID:                   1,
				Name:                 "Todo",
				ProjectID:            100,
				PrevID:               types.NullInt64{Valid: false},
				NextID:               types.NullInt64{Valid: false},
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
			expected: &models.Column{
				ID:                   1,
				Name:                 "Todo",
				ProjectID:            100,
				PrevID:               nil,
				NextID:               nil,
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
		},
		{
			name: "column with prev_id only",
			input: types.Column{
				ID:                   2,
				Name:                 "In Progress",
				ProjectID:            100,
				PrevID:               types.NullInt64{Int64: 1, Valid: true},
				NextID:               types.NullInt64{Valid: false},
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: true,
			},
			expected: &models.Column{
				ID:                   2,
				Name:                 "In Progress",
				ProjectID:            100,
				PrevID:               new(1),
				NextID:               nil,
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: true,
			},
		},
		{
			name: "column with next_id only",
			input: types.Column{
				ID:                   3,
				Name:                 "Todo",
				ProjectID:            100,
				PrevID:               types.NullInt64{Valid: false},
				NextID:               types.NullInt64{Int64: 4, Valid: true},
				HoldsReadyTasks:      true,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
			expected: &models.Column{
				ID:                   3,
				Name:                 "Todo",
				ProjectID:            100,
				PrevID:               nil,
				NextID:               new(4),
				HoldsReadyTasks:      true,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
		},
		{
			name: "column with both prev_id and next_id",
			input: types.Column{
				ID:                   4,
				Name:                 "In Progress",
				ProjectID:            100,
				PrevID:               types.NullInt64{Int64: 3, Valid: true},
				NextID:               types.NullInt64{Int64: 5, Valid: true},
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: true,
			},
			expected: &models.Column{
				ID:                   4,
				Name:                 "In Progress",
				ProjectID:            100,
				PrevID:               new(3),
				NextID:               new(5),
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: true,
			},
		},
		{
			name: "completed column with holds_completed_tasks flag",
			input: types.Column{
				ID:                   5,
				Name:                 "Done",
				ProjectID:            100,
				PrevID:               types.NullInt64{Int64: 4, Valid: true},
				NextID:               types.NullInt64{Valid: false},
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  true,
				HoldsInProgressTasks: false,
			},
			expected: &models.Column{
				ID:                   5,
				Name:                 "Done",
				ProjectID:            100,
				PrevID:               new(4),
				NextID:               nil,
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  true,
				HoldsInProgressTasks: false,
			},
		},
		{
			name: "column with all flags true",
			input: types.Column{
				ID:                   6,
				Name:                 "Multi-Flag Column",
				ProjectID:            100,
				PrevID:               types.NullInt64{Int64: 5, Valid: true},
				NextID:               types.NullInt64{Int64: 7, Valid: true},
				HoldsReadyTasks:      true,
				HoldsCompletedTasks:  true,
				HoldsInProgressTasks: true,
			},
			expected: &models.Column{
				ID:                   6,
				Name:                 "Multi-Flag Column",
				ProjectID:            100,
				PrevID:               new(5),
				NextID:               new(7),
				HoldsReadyTasks:      true,
				HoldsCompletedTasks:  true,
				HoldsInProgressTasks: true,
			},
		},
		{
			name: "empty name column",
			input: types.Column{
				ID:                   7,
				Name:                 "",
				ProjectID:            100,
				PrevID:               types.NullInt64{Valid: false},
				NextID:               types.NullInt64{Valid: false},
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
			expected: &models.Column{
				ID:                   7,
				Name:                 "",
				ProjectID:            100,
				PrevID:               nil,
				NextID:               nil,
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
		},
		{
			name: "column with max int64 values",
			input: types.Column{
				ID:                   math.MaxInt64,
				Name:                 "Max Value Column",
				ProjectID:            math.MaxInt64,
				PrevID:               types.NullInt64{Int64: math.MaxInt64 - 1, Valid: true},
				NextID:               types.NullInt64{Valid: false},
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
			expected: &models.Column{
				ID:                   int(math.MaxInt64),
				Name:                 "Max Value Column",
				ProjectID:            int(math.MaxInt64),
				PrevID:               new(int(math.MaxInt64 - 1)),
				NextID:               nil,
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
		},
		{
			name: "column with zero IDs",
			input: types.Column{
				ID:                   0,
				Name:                 "Zero ID Column",
				ProjectID:            0,
				PrevID:               types.NullInt64{Int64: 0, Valid: true},
				NextID:               types.NullInt64{Int64: 0, Valid: true},
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
			expected: &models.Column{
				ID:                   0,
				Name:                 "Zero ID Column",
				ProjectID:            0,
				PrevID:               new(0),
				NextID:               new(0),
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ColumnToModel(tt.input)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.ProjectID, result.ProjectID)
			assert.True(t, intPtrEqual(result.PrevID, tt.expected.PrevID))
			assert.True(t, intPtrEqual(result.NextID, tt.expected.NextID))
			assert.Equal(t, tt.expected.HoldsReadyTasks, result.HoldsReadyTasks)
			assert.Equal(t, tt.expected.HoldsCompletedTasks, result.HoldsCompletedTasks)
			assert.Equal(t, tt.expected.HoldsInProgressTasks, result.HoldsInProgressTasks)
		})
	}
}

func TestColumnFromIDRowToModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    types.GetColumnByIDRow
		expected *models.Column
	}{
		{
			name: "basic column from ID row",
			input: types.GetColumnByIDRow{
				ID:                   10,
				Name:                 "Backlog",
				ProjectID:            200,
				PrevID:               types.NullInt64{Valid: false},
				NextID:               types.NullInt64{Valid: false},
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
			expected: &models.Column{
				ID:                   10,
				Name:                 "Backlog",
				ProjectID:            200,
				PrevID:               nil,
				NextID:               nil,
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
		},
		{
			name: "column with linked list pointers",
			input: types.GetColumnByIDRow{
				ID:                   20,
				Name:                 "Review",
				ProjectID:            200,
				PrevID:               types.NullInt64{Int64: 19, Valid: true},
				NextID:               types.NullInt64{Int64: 21, Valid: true},
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
			expected: &models.Column{
				ID:                   20,
				Name:                 "Review",
				ProjectID:            200,
				PrevID:               new(19),
				NextID:               new(21),
				HoldsReadyTasks:      false,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
		},
		{
			name: "ready tasks column",
			input: types.GetColumnByIDRow{
				ID:                   30,
				Name:                 "Ready for Work",
				ProjectID:            200,
				PrevID:               types.NullInt64{Int64: 29, Valid: true},
				NextID:               types.NullInt64{Int64: 31, Valid: true},
				HoldsReadyTasks:      true,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
			expected: &models.Column{
				ID:                   30,
				Name:                 "Ready for Work",
				ProjectID:            200,
				PrevID:               new(29),
				NextID:               new(31),
				HoldsReadyTasks:      true,
				HoldsCompletedTasks:  false,
				HoldsInProgressTasks: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ColumnFromIDRowToModel(tt.input)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.ProjectID, result.ProjectID)
			assert.True(t, intPtrEqual(result.PrevID, tt.expected.PrevID))
			assert.True(t, intPtrEqual(result.NextID, tt.expected.NextID))
			assert.Equal(t, tt.expected.HoldsReadyTasks, result.HoldsReadyTasks)
			assert.Equal(t, tt.expected.HoldsCompletedTasks, result.HoldsCompletedTasks)
			assert.Equal(t, tt.expected.HoldsInProgressTasks, result.HoldsInProgressTasks)
		})
	}
}

func TestColumnsFromRowsToModels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []types.GetColumnsByProjectRow
		expected []*models.Column
	}{
		{
			name:     "empty slice",
			input:    []types.GetColumnsByProjectRow{},
			expected: []*models.Column{},
		},
		{
			name: "single column",
			input: []types.GetColumnsByProjectRow{
				{
					ID:                   1,
					Name:                 "Todo",
					ProjectID:            100,
					PrevID:               types.NullInt64{Valid: false},
					NextID:               types.NullInt64{Valid: false},
					HoldsReadyTasks:      true,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
			},
			expected: []*models.Column{
				{
					ID:                   1,
					Name:                 "Todo",
					ProjectID:            100,
					PrevID:               nil,
					NextID:               nil,
					HoldsReadyTasks:      true,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
			},
		},
		{
			name: "multiple columns forming linked list",
			input: []types.GetColumnsByProjectRow{
				{
					ID:                   1,
					Name:                 "Todo",
					ProjectID:            100,
					PrevID:               types.NullInt64{Valid: false},
					NextID:               types.NullInt64{Int64: 2, Valid: true},
					HoldsReadyTasks:      true,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   2,
					Name:                 "In Progress",
					ProjectID:            100,
					PrevID:               types.NullInt64{Int64: 1, Valid: true},
					NextID:               types.NullInt64{Int64: 3, Valid: true},
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: true,
				},
				{
					ID:                   3,
					Name:                 "Done",
					ProjectID:            100,
					PrevID:               types.NullInt64{Int64: 2, Valid: true},
					NextID:               types.NullInt64{Valid: false},
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  true,
					HoldsInProgressTasks: false,
				},
			},
			expected: []*models.Column{
				{
					ID:                   1,
					Name:                 "Todo",
					ProjectID:            100,
					PrevID:               nil,
					NextID:               new(2),
					HoldsReadyTasks:      true,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   2,
					Name:                 "In Progress",
					ProjectID:            100,
					PrevID:               new(1),
					NextID:               new(3),
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: true,
				},
				{
					ID:                   3,
					Name:                 "Done",
					ProjectID:            100,
					PrevID:               new(2),
					NextID:               nil,
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  true,
					HoldsInProgressTasks: false,
				},
			},
		},
		{
			name: "unlinked columns (no prev/next IDs)",
			input: []types.GetColumnsByProjectRow{
				{
					ID:                   10,
					Name:                 "Column A",
					ProjectID:            200,
					PrevID:               types.NullInt64{Valid: false},
					NextID:               types.NullInt64{Valid: false},
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   11,
					Name:                 "Column B",
					ProjectID:            200,
					PrevID:               types.NullInt64{Valid: false},
					NextID:               types.NullInt64{Valid: false},
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
			},
			expected: []*models.Column{
				{
					ID:                   10,
					Name:                 "Column A",
					ProjectID:            200,
					PrevID:               nil,
					NextID:               nil,
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   11,
					Name:                 "Column B",
					ProjectID:            200,
					PrevID:               nil,
					NextID:               nil,
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
			},
		},
		{
			name: "columns with various flag combinations",
			input: []types.GetColumnsByProjectRow{
				{
					ID:                   20,
					Name:                 "Backlog",
					ProjectID:            300,
					PrevID:               types.NullInt64{Valid: false},
					NextID:               types.NullInt64{Int64: 21, Valid: true},
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   21,
					Name:                 "Ready",
					ProjectID:            300,
					PrevID:               types.NullInt64{Int64: 20, Valid: true},
					NextID:               types.NullInt64{Int64: 22, Valid: true},
					HoldsReadyTasks:      true,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   22,
					Name:                 "Working",
					ProjectID:            300,
					PrevID:               types.NullInt64{Int64: 21, Valid: true},
					NextID:               types.NullInt64{Int64: 23, Valid: true},
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: true,
				},
				{
					ID:                   23,
					Name:                 "Completed",
					ProjectID:            300,
					PrevID:               types.NullInt64{Int64: 22, Valid: true},
					NextID:               types.NullInt64{Valid: false},
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  true,
					HoldsInProgressTasks: false,
				},
			},
			expected: []*models.Column{
				{
					ID:                   20,
					Name:                 "Backlog",
					ProjectID:            300,
					PrevID:               nil,
					NextID:               new(21),
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   21,
					Name:                 "Ready",
					ProjectID:            300,
					PrevID:               new(20),
					NextID:               new(22),
					HoldsReadyTasks:      true,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   22,
					Name:                 "Working",
					ProjectID:            300,
					PrevID:               new(21),
					NextID:               new(23),
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: true,
				},
				{
					ID:                   23,
					Name:                 "Completed",
					ProjectID:            300,
					PrevID:               new(22),
					NextID:               nil,
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  true,
					HoldsInProgressTasks: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ColumnsFromRowsToModels(tt.input)

			require.Len(t, result, len(tt.expected))

			for i, expected := range tt.expected {
				got := result[i]

				assert.Equal(t, expected.ID, got.ID)
				assert.Equal(t, expected.Name, got.Name)
				assert.Equal(t, expected.ProjectID, got.ProjectID)
				assert.True(t, intPtrEqual(got.PrevID, expected.PrevID))
				assert.True(t, intPtrEqual(got.NextID, expected.NextID))
				assert.Equal(t, expected.HoldsReadyTasks, got.HoldsReadyTasks)
				assert.Equal(t, expected.HoldsCompletedTasks, got.HoldsCompletedTasks)
				assert.Equal(t, expected.HoldsInProgressTasks, got.HoldsInProgressTasks)
			}
		})
	}
}

// intPtrEqual compares two *int pointers for equality
func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
