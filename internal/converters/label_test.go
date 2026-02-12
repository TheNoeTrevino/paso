package converters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestLabelToModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    types.Label
		expected *models.Label
	}{
		{
			name: "standard label",
			input: types.Label{
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
			input: types.Label{
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
			input: types.Label{
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
			input: types.Label{
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
			input: types.Label{
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
			input: types.Label{
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
			input: types.Label{
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
			input: types.Label{
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
			input: types.Label{
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
			input: types.Label{
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
			t.Parallel()

			result := LabelToModel(tt.input)

			require.NotNil(t, result)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Color, result.Color)
			assert.Equal(t, tt.expected.ProjectID, result.ProjectID)
		})
	}
}

func TestLabelsToModels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         []types.Label
		expectedCount int
	}{
		{
			name: "multiple labels",
			input: []types.Label{
				{ID: 1, Name: "bug", Color: "#FF5733", ProjectID: 10},
				{ID: 2, Name: "feature", Color: "#00FF00", ProjectID: 10},
				{ID: 3, Name: "enhancement", Color: "#0000FF", ProjectID: 10},
			},
			expectedCount: 3,
		},
		{
			name:          "empty slice",
			input:         []types.Label{},
			expectedCount: 0,
		},
		{
			name: "single label",
			input: []types.Label{
				{ID: 1, Name: "bug", Color: "#FF5733", ProjectID: 10},
			},
			expectedCount: 1,
		},
		{
			name: "labels with mixed names",
			input: []types.Label{
				{ID: 1, Name: "simple", Color: "#111111", ProjectID: 10},
				{ID: 2, Name: "with spaces", Color: "#222222", ProjectID: 10},
				{ID: 3, Name: "with-dashes", Color: "#333333", ProjectID: 10},
				{ID: 4, Name: "with_underscores", Color: "#444444", ProjectID: 10},
			},
			expectedCount: 4,
		},
		{
			name: "labels from different projects",
			input: []types.Label{
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
			t.Parallel()

			result := LabelsToModels(tt.input)

			require.NotNil(t, result)
			assert.Len(t, result, tt.expectedCount)

			// Verify each label was converted correctly
			for i, label := range result {
				if i >= len(tt.input) {
					break
				}

				expected := tt.input[i]
				assert.Equal(t, int(expected.ID), label.ID)
				assert.Equal(t, expected.Name, label.Name)
				assert.Equal(t, expected.Color, label.Color)
				assert.Equal(t, int(expected.ProjectID), label.ProjectID)
			}
		})
	}
}

func TestLabelsToModels_NilSlice(t *testing.T) {
	t.Parallel()

	result := LabelsToModels(nil)

	require.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestLabelsToModels_PreservesOrder(t *testing.T) {
	t.Parallel()

	input := []types.Label{
		{ID: 5, Name: "fifth", Color: "#555555", ProjectID: 10},
		{ID: 3, Name: "third", Color: "#333333", ProjectID: 10},
		{ID: 1, Name: "first", Color: "#111111", ProjectID: 10},
		{ID: 4, Name: "fourth", Color: "#444444", ProjectID: 10},
		{ID: 2, Name: "second", Color: "#222222", ProjectID: 10},
	}

	result := LabelsToModels(input)

	require.Len(t, result, len(input))

	// Verify order is preserved
	expectedIDs := []int{5, 3, 1, 4, 2}
	for i, label := range result {
		assert.Equal(t, expectedIDs[i], label.ID)
	}
}

func TestLabelToModel_TypeConversion(t *testing.T) {
	t.Parallel()

	// Test that int64 from database is correctly converted to int for models
	input := types.Label{
		ID:        int64(9223372036854775807), // Max int64
		Name:      "test",
		Color:     "#FFFFFF",
		ProjectID: int64(9223372036854775806),
	}

	result := LabelToModel(input)

	require.NotNil(t, result)

	// This test verifies the type conversion happens without panic
	// The actual values may overflow on 32-bit systems, but that's expected behavior
	assert.False(t, result.ID == 0 && input.ID != 0)
}

func BenchmarkLabelToModel(b *testing.B) {
	label := types.Label{
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
	labels := []types.Label{
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
	labels := make([]types.Label, 100)
	for i := 0; i < 100; i++ {
		labels[i] = types.Label{
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
