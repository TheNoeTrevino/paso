package converters

import (
	"testing"

	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/models"
)

// ============================================================================
// TEST CASES - LabelToModel
// ============================================================================

func TestLabelToModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    generated.Label
		expected *models.Label
	}{
		{
			name: "standard label",
			input: generated.Label{
				ID:        1,
				Name:      "bug",
				Color:     "#FF5733",
				ProjectID: 10,
			},
			expected: &models.Label{
				ID:        1,
				Name:      "bug",
				Color:     "#FF5733",
				ProjectID: 10,
			},
		},
		{
			name: "label with spaces in name",
			input: generated.Label{
				ID:        2,
				Name:      "needs review",
				Color:     "#00FF00",
				ProjectID: 10,
			},
			expected: &models.Label{
				ID:        2,
				Name:      "needs review",
				Color:     "#00FF00",
				ProjectID: 10,
			},
		},
		{
			name: "label with special characters",
			input: generated.Label{
				ID:        3,
				Name:      "p1-urgent!",
				Color:     "#FF0000",
				ProjectID: 10,
			},
			expected: &models.Label{
				ID:        3,
				Name:      "p1-urgent!",
				Color:     "#FF0000",
				ProjectID: 10,
			},
		},
		{
			name: "label with unicode characters",
			input: generated.Label{
				ID:        4,
				Name:      "优先级高",
				Color:     "#FFFF00",
				ProjectID: 10,
			},
			expected: &models.Label{
				ID:        4,
				Name:      "优先级高",
				Color:     "#FFFF00",
				ProjectID: 10,
			},
		},
		{
			name: "label with emojis",
			input: generated.Label{
				ID:        5,
				Name:      "🐛 bug",
				Color:     "#FF5733",
				ProjectID: 10,
			},
			expected: &models.Label{
				ID:        5,
				Name:      "🐛 bug",
				Color:     "#FF5733",
				ProjectID: 10,
			},
		},
		{
			name: "label with very long name",
			input: generated.Label{
				ID:        6,
				Name:      "this-is-a-very-long-label-name-that-exceeds-typical-length-but-should-still-be-handled-correctly",
				Color:     "#00FF00",
				ProjectID: 10,
			},
			expected: &models.Label{
				ID:        6,
				Name:      "this-is-a-very-long-label-name-that-exceeds-typical-length-but-should-still-be-handled-correctly",
				Color:     "#00FF00",
				ProjectID: 10,
			},
		},
		{
			name: "label with empty name",
			input: generated.Label{
				ID:        7,
				Name:      "",
				Color:     "#000000",
				ProjectID: 10,
			},
			expected: &models.Label{
				ID:        7,
				Name:      "",
				Color:     "#000000",
				ProjectID: 10,
			},
		},
		{
			name: "label with uppercase color",
			input: generated.Label{
				ID:        8,
				Name:      "feature",
				Color:     "#ABCDEF",
				ProjectID: 10,
			},
			expected: &models.Label{
				ID:        8,
				Name:      "feature",
				Color:     "#ABCDEF",
				ProjectID: 10,
			},
		},
		{
			name: "label with lowercase color",
			input: generated.Label{
				ID:        9,
				Name:      "enhancement",
				Color:     "#abcdef",
				ProjectID: 10,
			},
			expected: &models.Label{
				ID:        9,
				Name:      "enhancement",
				Color:     "#abcdef",
				ProjectID: 10,
			},
		},
		{
			name: "label with large IDs",
			input: generated.Label{
				ID:        999999,
				Name:      "test",
				Color:     "#123456",
				ProjectID: 888888,
			},
			expected: &models.Label{
				ID:        999999,
				Name:      "test",
				Color:     "#123456",
				ProjectID: 888888,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := LabelToModel(tt.input)

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if result.ID != tt.expected.ID {
				t.Errorf("Expected ID %d, got %d", tt.expected.ID, result.ID)
			}

			if result.Name != tt.expected.Name {
				t.Errorf("Expected Name %q, got %q", tt.expected.Name, result.Name)
			}

			if result.Color != tt.expected.Color {
				t.Errorf("Expected Color %q, got %q", tt.expected.Color, result.Color)
			}

			if result.ProjectID != tt.expected.ProjectID {
				t.Errorf("Expected ProjectID %d, got %d", tt.expected.ProjectID, result.ProjectID)
			}
		})
	}
}

// ============================================================================
// TEST CASES - LabelsToModels
// ============================================================================

func TestLabelsToModels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         []generated.Label
		expectedCount int
	}{
		{
			name: "multiple labels",
			input: []generated.Label{
				{ID: 1, Name: "bug", Color: "#FF5733", ProjectID: 10},
				{ID: 2, Name: "feature", Color: "#00FF00", ProjectID: 10},
				{ID: 3, Name: "enhancement", Color: "#0000FF", ProjectID: 10},
			},
			expectedCount: 3,
		},
		{
			name:          "empty slice",
			input:         []generated.Label{},
			expectedCount: 0,
		},
		{
			name: "single label",
			input: []generated.Label{
				{ID: 1, Name: "bug", Color: "#FF5733", ProjectID: 10},
			},
			expectedCount: 1,
		},
		{
			name: "labels with mixed names",
			input: []generated.Label{
				{ID: 1, Name: "simple", Color: "#111111", ProjectID: 10},
				{ID: 2, Name: "with spaces", Color: "#222222", ProjectID: 10},
				{ID: 3, Name: "with-dashes", Color: "#333333", ProjectID: 10},
				{ID: 4, Name: "with_underscores", Color: "#444444", ProjectID: 10},
			},
			expectedCount: 4,
		},
		{
			name: "labels from different projects",
			input: []generated.Label{
				{ID: 1, Name: "bug", Color: "#FF5733", ProjectID: 10},
				{ID: 2, Name: "bug", Color: "#FF5733", ProjectID: 20},
				{ID: 3, Name: "feature", Color: "#00FF00", ProjectID: 30},
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := LabelsToModels(tt.input)

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d labels, got %d", tt.expectedCount, len(result))
			}

			// Verify each label was converted correctly
			for i, label := range result {
				if i >= len(tt.input) {
					break
				}

				expected := tt.input[i]
				if label.ID != int(expected.ID) {
					t.Errorf("Label %d: Expected ID %d, got %d", i, expected.ID, label.ID)
				}
				if label.Name != expected.Name {
					t.Errorf("Label %d: Expected Name %q, got %q", i, expected.Name, label.Name)
				}
				if label.Color != expected.Color {
					t.Errorf("Label %d: Expected Color %q, got %q", i, expected.Color, label.Color)
				}
				if label.ProjectID != int(expected.ProjectID) {
					t.Errorf("Label %d: Expected ProjectID %d, got %d", i, expected.ProjectID, label.ProjectID)
				}
			}
		})
	}
}

func TestLabelsToModels_NilSlice(t *testing.T) {
	t.Parallel()

	result := LabelsToModels(nil)

	if result == nil {
		t.Fatal("Expected non-nil result for nil input")
	}

	if len(result) != 0 {
		t.Errorf("Expected empty slice for nil input, got length %d", len(result))
	}
}

func TestLabelsToModels_PreservesOrder(t *testing.T) {
	t.Parallel()

	input := []generated.Label{
		{ID: 5, Name: "fifth", Color: "#555555", ProjectID: 10},
		{ID: 3, Name: "third", Color: "#333333", ProjectID: 10},
		{ID: 1, Name: "first", Color: "#111111", ProjectID: 10},
		{ID: 4, Name: "fourth", Color: "#444444", ProjectID: 10},
		{ID: 2, Name: "second", Color: "#222222", ProjectID: 10},
	}

	result := LabelsToModels(input)

	if len(result) != len(input) {
		t.Fatalf("Expected %d labels, got %d", len(input), len(result))
	}

	// Verify order is preserved
	expectedIDs := []int{5, 3, 1, 4, 2}
	for i, label := range result {
		if label.ID != expectedIDs[i] {
			t.Errorf("Position %d: Expected ID %d, got %d", i, expectedIDs[i], label.ID)
		}
	}
}

// ============================================================================
// TEST CASES - Type Conversion
// ============================================================================

func TestLabelToModel_TypeConversion(t *testing.T) {
	t.Parallel()

	// Test that int64 from database is correctly converted to int for models
	input := generated.Label{
		ID:        int64(9223372036854775807), // Max int64
		Name:      "test",
		Color:     "#FFFFFF",
		ProjectID: int64(9223372036854775806),
	}

	result := LabelToModel(input)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// This test verifies the type conversion happens without panic
	// The actual values may overflow on 32-bit systems, but that's expected behavior
	if result.ID == 0 && input.ID != 0 {
		t.Error("ID conversion resulted in zero when input was non-zero")
	}
}

// ============================================================================
// BENCHMARK TESTS
// ============================================================================

func BenchmarkLabelToModel(b *testing.B) {
	label := generated.Label{
		ID:        1,
		Name:      "bug",
		Color:     "#FF5733",
		ProjectID: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = LabelToModel(label)
	}
}

func BenchmarkLabelsToModels(b *testing.B) {
	labels := []generated.Label{
		{ID: 1, Name: "bug", Color: "#FF5733", ProjectID: 10},
		{ID: 2, Name: "feature", Color: "#00FF00", ProjectID: 10},
		{ID: 3, Name: "enhancement", Color: "#0000FF", ProjectID: 10},
		{ID: 4, Name: "documentation", Color: "#FFFF00", ProjectID: 10},
		{ID: 5, Name: "help wanted", Color: "#FF00FF", ProjectID: 10},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = LabelsToModels(labels)
	}
}

func BenchmarkLabelsToModels_Large(b *testing.B) {
	// Create a large slice of labels
	labels := make([]generated.Label, 100)
	for i := 0; i < 100; i++ {
		labels[i] = generated.Label{
			ID:        int64(i + 1),
			Name:      "label-" + string(rune('0'+i%10)),
			Color:     "#FFFFFF",
			ProjectID: 10,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = LabelsToModels(labels)
	}
}
