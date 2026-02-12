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
				{ID: 2, Name: "In Progress", NextID: &next3ID, PrevID: intPtr(1)},
				{ID: 3, Name: "Done", PrevID: intPtr(2)},
			},
			currentColumnID: 1,
			expected:        "In Progress",
		},
		{
			name: "last column has no next",
			columns: []*models.Column{
				{ID: 1, Name: "Todo", NextID: &next2ID},
				{ID: 2, Name: "Done", PrevID: intPtr(1)},
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
				{ID: 2, Name: "In Progress", NextID: &next3ID, PrevID: intPtr(1)},
				{ID: 3, Name: "Done", PrevID: intPtr(2)},
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
				{ID: 2, Name: "Done", PrevID: intPtr(1)},
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
				{ID: 2, Name: "In Progress", NextID: &next3ID, PrevID: intPtr(1)},
				{ID: 3, Name: "Done", PrevID: intPtr(2)},
			},
			currentColumnID: 2,
			expected:        "Todo",
		},
		{
			name: "first column has no previous",
			columns: []*models.Column{
				{ID: 1, Name: "Todo", NextID: &next2ID},
				{ID: 2, Name: "Done", PrevID: intPtr(1)},
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
				{ID: 2, Name: "In Progress", NextID: &next3ID, PrevID: intPtr(1)},
				{ID: 3, Name: "Done", PrevID: intPtr(2)},
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
				{ID: 2, Name: "Done", PrevID: intPtr(1)},
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

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}
