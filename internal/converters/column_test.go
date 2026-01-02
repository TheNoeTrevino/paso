package converters

import (
	"math"
	"testing"

	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/models"
)

// ============================================================================
// TEST CASES - ColumnToModel
// ============================================================================

func TestColumnToModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    generated.Column
		expected *models.Column
	}{
		{
			name: "basic column with all flags false",
			input: generated.Column{
				ID:                   1,
				Name:                 "Todo",
				ProjectID:            100,
				PrevID:               nil,
				NextID:               nil,
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
			input: generated.Column{
				ID:                   2,
				Name:                 "In Progress",
				ProjectID:            100,
				PrevID:               int64(1),
				NextID:               nil,
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
			input: generated.Column{
				ID:                   3,
				Name:                 "Todo",
				ProjectID:            100,
				PrevID:               nil,
				NextID:               int64(4),
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
			input: generated.Column{
				ID:                   4,
				Name:                 "In Progress",
				ProjectID:            100,
				PrevID:               int64(3),
				NextID:               int64(5),
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
			input: generated.Column{
				ID:                   5,
				Name:                 "Done",
				ProjectID:            100,
				PrevID:               int64(4),
				NextID:               nil,
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
			input: generated.Column{
				ID:                   6,
				Name:                 "Multi-Flag Column",
				ProjectID:            100,
				PrevID:               int64(5),
				NextID:               int64(7),
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
			input: generated.Column{
				ID:                   7,
				Name:                 "",
				ProjectID:            100,
				PrevID:               nil,
				NextID:               nil,
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
			input: generated.Column{
				ID:                   math.MaxInt64,
				Name:                 "Max Value Column",
				ProjectID:            math.MaxInt64,
				PrevID:               int64(math.MaxInt64 - 1),
				NextID:               nil,
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
			input: generated.Column{
				ID:                   0,
				Name:                 "Zero ID Column",
				ProjectID:            0,
				PrevID:               int64(0),
				NextID:               int64(0),
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
		input    generated.GetColumnByIDRow
		expected *models.Column
	}{
		{
			name: "basic column from ID row",
			input: generated.GetColumnByIDRow{
				ID:                   10,
				Name:                 "Backlog",
				ProjectID:            200,
				PrevID:               nil,
				NextID:               nil,
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
			input: generated.GetColumnByIDRow{
				ID:                   20,
				Name:                 "Review",
				ProjectID:            200,
				PrevID:               int64(19),
				NextID:               int64(21),
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
			input: generated.GetColumnByIDRow{
				ID:                   30,
				Name:                 "Ready for Work",
				ProjectID:            200,
				PrevID:               int64(29),
				NextID:               int64(31),
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
		input    []generated.GetColumnsByProjectRow
		expected []*models.Column
	}{
		{
			name:     "empty slice",
			input:    []generated.GetColumnsByProjectRow{},
			expected: []*models.Column{},
		},
		{
			name: "single column",
			input: []generated.GetColumnsByProjectRow{
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
			input: []generated.GetColumnsByProjectRow{
				{
					ID:                   1,
					Name:                 "Todo",
					ProjectID:            100,
					PrevID:               nil,
					NextID:               int64(2),
					HoldsReadyTasks:      true,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   2,
					Name:                 "In Progress",
					ProjectID:            100,
					PrevID:               int64(1),
					NextID:               int64(3),
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: true,
				},
				{
					ID:                   3,
					Name:                 "Done",
					ProjectID:            100,
					PrevID:               int64(2),
					NextID:               nil,
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
			input: []generated.GetColumnsByProjectRow{
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
			input: []generated.GetColumnsByProjectRow{
				{
					ID:                   20,
					Name:                 "Backlog",
					ProjectID:            300,
					PrevID:               nil,
					NextID:               int64(21),
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   21,
					Name:                 "Ready",
					ProjectID:            300,
					PrevID:               int64(20),
					NextID:               int64(22),
					HoldsReadyTasks:      true,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: false,
				},
				{
					ID:                   22,
					Name:                 "Working",
					ProjectID:            300,
					PrevID:               int64(21),
					NextID:               int64(23),
					HoldsReadyTasks:      false,
					HoldsCompletedTasks:  false,
					HoldsInProgressTasks: true,
				},
				{
					ID:                   23,
					Name:                 "Completed",
					ProjectID:            300,
					PrevID:               int64(22),
					NextID:               nil,
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
