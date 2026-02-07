package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestFindColumnByName_Found(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo", ProjectID: 1},
		{ID: 2, Name: "In Progress", ProjectID: 1},
		{ID: 3, Name: "Done", ProjectID: 1},
	}

	tests := []struct {
		name       string
		searchName string
		expectedID int
	}{
		{"exact match", "Todo", 1},
		{"exact match with spaces", "In Progress", 2},
		{"lowercase", "todo", 1},
		{"uppercase", "TODO", 1},
		{"mixed case", "ToDo", 1},
		{"lowercase with spaces", "in progress", 2},
		{"uppercase with spaces", "IN PROGRESS", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col, err := FindColumnByName(columns, tt.searchName)
			assert.NoError(t, err, "FindColumnByName should not return error for: %s", tt.searchName)
			if col != nil && col.ID != tt.expectedID {
				t.Errorf("Expected column ID %d, got %d", tt.expectedID, col.ID)
			}
		})
	}
}

func TestFindColumnByName_NotFound(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo", ProjectID: 1},
		{ID: 2, Name: "In Progress", ProjectID: 1},
		{ID: 3, Name: "Done", ProjectID: 1},
	}

	tests := []string{
		"Nonexistent",
		"Doing",
		"",
		"Tod", // partial match should not work
		"Todoo",
	}

	for _, searchName := range tests {
		t.Run(searchName, func(t *testing.T) {
			_, err := FindColumnByName(columns, searchName)
			if err == nil {
				t.Errorf("Expected error for '%s', got nil", searchName)
			}
		})
	}
}

func TestFindColumnByName_EmptyList(t *testing.T) {
	columns := []*models.Column{}

	_, err := FindColumnByName(columns, "Todo")
	if err == nil {
		t.Error("Expected error for empty column list, got nil")
	}
}

func TestFormatAvailableColumns(t *testing.T) {
	tests := []struct {
		name     string
		columns  []*models.Column
		expected string
	}{
		{
			name: "multiple columns",
			columns: []*models.Column{
				{ID: 1, Name: "Todo"},
				{ID: 2, Name: "In Progress"},
				{ID: 3, Name: "Done"},
			},
			expected: "Todo, In Progress, Done",
		},
		{
			name: "single column",
			columns: []*models.Column{
				{ID: 1, Name: "Backlog"},
			},
			expected: "Backlog",
		},
		{
			name:     "empty list",
			columns:  []*models.Column{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatAvailableColumns(tt.columns)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetCurrentColumnName(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo", ProjectID: 1},
		{ID: 2, Name: "In Progress", ProjectID: 1},
		{ID: 3, Name: "Done", ProjectID: 1},
	}

	tests := []struct {
		name     string
		columnID int
		expected string
	}{
		{"first column", 1, "Todo"},
		{"middle column", 2, "In Progress"},
		{"last column", 3, "Done"},
		{"non-existent column", 999, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCurrentColumnName(columns, tt.columnID)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetCurrentColumnName_EmptyList(t *testing.T) {
	columns := []*models.Column{}

	result := GetCurrentColumnName(columns, 1)
	if result != "Unknown" {
		t.Errorf("Expected 'Unknown' for empty list, got '%s'", result)
	}
}
