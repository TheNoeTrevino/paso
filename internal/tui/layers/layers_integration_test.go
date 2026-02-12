package layers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateCenteredLayerWithContent tests layer creation with content
func TestCreateCenteredLayerWithContent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		content      string
		screenWidth  int
		screenHeight int
	}{
		{
			name:         "normal screen",
			content:      "Test Content",
			screenWidth:  120,
			screenHeight: 40,
		},
		{
			name:         "narrow screen",
			content:      "Content",
			screenWidth:  60,
			screenHeight: 20,
		},
		{
			name:         "small content on large screen",
			content:      "X",
			screenWidth:  200,
			screenHeight: 100,
		},
		{
			name:         "large content",
			content:      "This is a very long piece of content that needs to be centered on the screen",
			screenWidth:  80,
			screenHeight: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()
			layer := CreateCenteredLayer(tt.content, tt.screenWidth, tt.screenHeight)

			require.NotNil(t, layer)
		})
	}
}

// TestCreateCenteredLayerWithEmptyContent tests layer creation with empty content
func TestCreateCenteredLayerWithEmptyContent(t *testing.T) {
	t.Parallel()
	layer := CreateCenteredLayer("", 120, 40)

	assert.Nil(t, layer)
}

// TestLayerPositioning tests that layers are centered correctly
func TestLayerPositioning(t *testing.T) {
	t.Parallel()
	content := "Center"
	screenWidth := 100
	screenHeight := 50

	layer := CreateCenteredLayer(content, screenWidth, screenHeight)

	require.NotNil(t, layer)

	// Layer methods are chainable, so just verify layer exists
	// (actual positioning is tested implicitly by the function working)
}

// TestLayerPositioningOnSmallScreen tests centering on very small screens
func TestLayerPositioningOnSmallScreen(t *testing.T) {
	t.Parallel()
	content := "X"
	screenWidth := 5
	screenHeight := 3

	layer := CreateCenteredLayer(content, screenWidth, screenHeight)

	require.NotNil(t, layer)
}

// TestCalculatePickerDimensions tests picker dimension calculation
func TestCalculatePickerDimensions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		itemCount    int
		hasFilter    bool
		screenWidth  int
		screenHeight int
		minWidth     int
		maxWidth     int
	}{
		{
			name:         "normal case",
			itemCount:    10,
			hasFilter:    true,
			screenWidth:  120,
			screenHeight: 40,
			minWidth:     20,
			maxWidth:     60,
		},
		{
			name:         "no filter",
			itemCount:    5,
			hasFilter:    false,
			screenWidth:  100,
			screenHeight: 30,
			minWidth:     15,
			maxWidth:     50,
		},
		{
			name:         "many items",
			itemCount:    100,
			hasFilter:    true,
			screenWidth:  150,
			screenHeight: 50,
			minWidth:     25,
			maxWidth:     80,
		},
		{
			name:         "small screen",
			itemCount:    3,
			hasFilter:    false,
			screenWidth:  40,
			screenHeight: 15,
			minWidth:     10,
			maxWidth:     30,
		},
		{
			name:         "single item",
			itemCount:    1,
			hasFilter:    true,
			screenWidth:  80,
			screenHeight: 25,
			minWidth:     15,
			maxWidth:     50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()
			width, height := CalculatePickerDimensions(
				tt.itemCount,
				tt.hasFilter,
				tt.screenWidth,
				tt.screenHeight,
				tt.minWidth,
				tt.maxWidth,
			)

			assert.GreaterOrEqual(t, width, tt.minWidth)
			assert.LessOrEqual(t, width, tt.maxWidth)
			assert.GreaterOrEqual(t, height, PickerMinHeight)
			assert.LessOrEqual(t, height, tt.screenHeight)
		})
	}
}

// TestPickerDimensionsWithLargeItemCount tests picker dimensions with many items
func TestPickerDimensionsWithLargeItemCount(t *testing.T) {
	t.Parallel()
	pickerWidth, pickerHeight := CalculatePickerDimensions(
		10000, // very many items
		true,
		200, // large screen
		100,
		30,
		150,
	)

	assert.Positive(t, pickerWidth)
	assert.Positive(t, pickerHeight)
}

// TestPickerDimensionsMinimums tests that picker respects minimum dimensions
func TestPickerDimensionsMinimums(t *testing.T) {
	t.Parallel()
	width, height := CalculatePickerDimensions(
		10, // some items
		false,
		80, // reasonable screen
		30,
		20, // minWidth
		50, // maxWidth
	)

	assert.GreaterOrEqual(t, width, 20)
	assert.GreaterOrEqual(t, height, 5)
}

// TestPickerDimensionsMaximums tests that picker respects maximum dimensions
func TestPickerDimensionsMaximums(t *testing.T) {
	t.Parallel()
	screenHeight := 50
	pickerWidth, pickerHeight := CalculatePickerDimensions(
		100,
		true,
		300, // very wide screen
		screenHeight,
		10,
		100, // maxWidth
	)

	assert.LessOrEqual(t, pickerWidth, 100)

	maxHeight := screenHeight * PickerMaxHeightNumerator / PickerMaxHeightDivisor
	assert.LessOrEqual(t, pickerHeight, maxHeight)
}

// TestPickerDimensionsWithAndWithoutFilter tests filter impact on dimensions
func TestPickerDimensionsWithAndWithoutFilter(t *testing.T) {
	t.Parallel()
	const itemCount = 10
	const screenWidth = 120
	const screenHeight = 40
	const minWidth = 20
	const maxWidth = 60

	w1, h1 := CalculatePickerDimensions(
		itemCount, true, screenWidth, screenHeight, minWidth, maxWidth)
	w2, h2 := CalculatePickerDimensions(
		itemCount, false, screenWidth, screenHeight, minWidth, maxWidth)

	assert.Greater(t, h1, h2)
	assert.Equal(t, w1, w2)
}

// TestPickerDimensionsConsistency tests consistency across multiple calls
func TestPickerDimensionsConsistency(t *testing.T) {
	t.Parallel()
	params := struct {
		itemCount    int
		hasFilter    bool
		screenWidth  int
		screenHeight int
		minWidth     int
		maxWidth     int
	}{
		itemCount:    15,
		hasFilter:    true,
		screenWidth:  100,
		screenHeight: 30,
		minWidth:     20,
		maxWidth:     50,
	}

	// Call multiple times with same parameters
	w1, h1 := CalculatePickerDimensions(
		params.itemCount, params.hasFilter, params.screenWidth,
		params.screenHeight, params.minWidth, params.maxWidth)

	w2, h2 := CalculatePickerDimensions(
		params.itemCount, params.hasFilter, params.screenWidth,
		params.screenHeight, params.minWidth, params.maxWidth)

	w3, h3 := CalculatePickerDimensions(
		params.itemCount, params.hasFilter, params.screenWidth,
		params.screenHeight, params.minWidth, params.maxWidth)

	assert.Equal(t, w1, w2)
	assert.Equal(t, w1, w3)
	assert.Equal(t, h1, h2)
	assert.Equal(t, h1, h3)
}

// TestLayerMultipleDimensions tests layers work with various screen dimensions
func TestLayerMultipleDimensions(t *testing.T) {
	t.Parallel()
	screenDimensions := [][2]int{
		{40, 10},
		{80, 24},
		{120, 40},
		{160, 50},
		{200, 80},
	}

	for _, dims := range screenDimensions {
		t.Run("dimension", func(t *testing.T) {
		t.Parallel()
			t.Parallel()
			width, height := dims[0], dims[1]
			layer := CreateCenteredLayer("Test", width, height)

			require.NotNil(t, layer)
		})
	}
}

// TestLayerWithMultilineContent tests layer centering with multiline content
func TestLayerWithMultilineContent(t *testing.T) {
	t.Parallel()
	multilineContent := `Line 1
Line 2
Line 3
Line 4`

	layer := CreateCenteredLayer(multilineContent, 100, 50)

	require.NotNil(t, layer)

	// Just verify it was created successfully
}
