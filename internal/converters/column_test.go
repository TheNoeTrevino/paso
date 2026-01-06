package converters

import (
	"math"
	"testing"

	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/models"
)

// ============================================================================
// TEST CASES - ColumnToModel
// ============================================================================

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
				PrevID:               intPtr(1),
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
				NextID:               intPtr(4),
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
				PrevID:               intPtr(3),
				NextID:               intPtr(5),
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
				PrevID:               intPtr(4),
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
				PrevID:               intPtr(5),
				NextID:               intPtr(7),
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
				PrevID:               intPtr(int(math.MaxInt64 - 1)),
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
				PrevID:               intPtr(0),
				NextID:               intPtr(0),
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

			// Verify all fields
			if result.ID != tt.expected.ID {
				t.Errorf("ID: got %d, want %d", result.ID, tt.expected.ID)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tt.expected.Name)
			}
			if result.ProjectID != tt.expected.ProjectID {
				t.Errorf("ProjectID: got %d, want %d", result.ProjectID, tt.expected.ProjectID)
			}
			if !intPtrEqual(result.PrevID, tt.expected.PrevID) {
				t.Errorf("PrevID: got %v, want %v", ptrToString(result.PrevID), ptrToString(tt.expected.PrevID))
			}
			if !intPtrEqual(result.NextID, tt.expected.NextID) {
				t.Errorf("NextID: got %v, want %v", ptrToString(result.NextID), ptrToString(tt.expected.NextID))
			}
			if result.HoldsReadyTasks != tt.expected.HoldsReadyTasks {
				t.Errorf("HoldsReadyTasks: got %v, want %v", result.HoldsReadyTasks, tt.expected.HoldsReadyTasks)
			}
			if result.HoldsCompletedTasks != tt.expected.HoldsCompletedTasks {
				t.Errorf("HoldsCompletedTasks: got %v, want %v", result.HoldsCompletedTasks, tt.expected.HoldsCompletedTasks)
			}
			if result.HoldsInProgressTasks != tt.expected.HoldsInProgressTasks {
				t.Errorf("HoldsInProgressTasks: got %v, want %v", result.HoldsInProgressTasks, tt.expected.HoldsInProgressTasks)
			}
		})
	}
}

// ============================================================================
// TEST CASES - ColumnFromIDRowToModel
// ============================================================================

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
				PrevID:               intPtr(19),
				NextID:               intPtr(21),
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
				PrevID:               intPtr(29),
				NextID:               intPtr(31),
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

			// Verify all fields
			if result.ID != tt.expected.ID {
				t.Errorf("ID: got %d, want %d", result.ID, tt.expected.ID)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tt.expected.Name)
			}
			if result.ProjectID != tt.expected.ProjectID {
				t.Errorf("ProjectID: got %d, want %d", result.ProjectID, tt.expected.ProjectID)
			}
			if !intPtrEqual(result.PrevID, tt.expected.PrevID) {
				t.Errorf("PrevID: got %v, want %v", ptrToString(result.PrevID), ptrToString(tt.expected.PrevID))
			}
			if !intPtrEqual(result.NextID, tt.expected.NextID) {
				t.Errorf("NextID: got %v, want %v", ptrToString(result.NextID), ptrToString(tt.expected.NextID))
			}
			if result.HoldsReadyTasks != tt.expected.HoldsReadyTasks {
				t.Errorf("HoldsReadyTasks: got %v, want %v", result.HoldsReadyTasks, tt.expected.HoldsReadyTasks)
			}
			if result.HoldsCompletedTasks != tt.expected.HoldsCompletedTasks {
				t.Errorf("HoldsCompletedTasks: got %v, want %v", result.HoldsCompletedTasks, tt.expected.HoldsCompletedTasks)
			}
			if result.HoldsInProgressTasks != tt.expected.HoldsInProgressTasks {
				t.Errorf("HoldsInProgressTasks: got %v, want %v", result.HoldsInProgressTasks, tt.expected.HoldsInProgressTasks)
			}
		})
	}
}

// ============================================================================
// TEST CASES - ColumnsFromRowsToModels
// ============================================================================

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
					NextID:               intPtr(2),
					HoldsReadyTasks:      true,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   2,
					Name:                 "In Progress",
					ProjectID:            100,
					PrevID:               intPtr(1),
					NextID:               intPtr(3),
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: true,
				},
				{
					ID:                   3,
					Name:                 "Done",
					ProjectID:            100,
					PrevID:               intPtr(2),
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
					NextID:               intPtr(21),
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   21,
					Name:                 "Ready",
					ProjectID:            300,
					PrevID:               intPtr(20),
					NextID:               intPtr(22),
					HoldsReadyTasks:      true,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   22,
					Name:                 "Working",
					ProjectID:            300,
					PrevID:               intPtr(21),
					NextID:               intPtr(23),
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: true,
				},
				{
					ID:                   23,
					Name:                 "Completed",
					ProjectID:            300,
					PrevID:               intPtr(22),
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

			// Verify length
			if len(result) != len(tt.expected) {
				t.Fatalf("Length: got %d, want %d", len(result), len(tt.expected))
			}

			// Verify each column
			for i, expected := range tt.expected {
				got := result[i]

				if got.ID != expected.ID {
					t.Errorf("Column %d - ID: got %d, want %d", i, got.ID, expected.ID)
				}
				if got.Name != expected.Name {
					t.Errorf("Column %d - Name: got %q, want %q", i, got.Name, expected.Name)
				}
				if got.ProjectID != expected.ProjectID {
					t.Errorf("Column %d - ProjectID: got %d, want %d", i, got.ProjectID, expected.ProjectID)
				}
				if !intPtrEqual(got.PrevID, expected.PrevID) {
					t.Errorf("Column %d - PrevID: got %v, want %v", i, ptrToString(got.PrevID), ptrToString(expected.PrevID))
				}
				if !intPtrEqual(got.NextID, expected.NextID) {
					t.Errorf("Column %d - NextID: got %v, want %v", i, ptrToString(got.NextID), ptrToString(expected.NextID))
				}
				if got.HoldsReadyTasks != expected.HoldsReadyTasks {
					t.Errorf("Column %d - HoldsReadyTasks: got %v, want %v", i, got.HoldsReadyTasks, expected.HoldsReadyTasks)
				}
				if got.HoldsCompletedTasks != expected.HoldsCompletedTasks {
					t.Errorf("Column %d - HoldsCompletedTasks: got %v, want %v", i, got.HoldsCompletedTasks, expected.HoldsCompletedTasks)
				}
				if got.HoldsInProgressTasks != expected.HoldsInProgressTasks {
					t.Errorf("Column %d - HoldsInProgressTasks: got %v, want %v", i, got.HoldsInProgressTasks, expected.HoldsInProgressTasks)
				}
			}
		})
	}
}

// ============================================================================
// TEST HELPERS
// ============================================================================

// intPtr returns a pointer to an int value
func intPtr(i int) *int {
	return &i
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

// ptrToString converts a pointer to a string representation for error messages
func ptrToString(p *int) string {
	if p == nil {
		return "nil"
	}
	return string(rune(*p + '0'))
}
