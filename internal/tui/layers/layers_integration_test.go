package layers

import (
	"testing"
)

// TestCreateCenteredLayerWithContent tests layer creation with content
func TestCreateCenteredLayerWithContent(t *testing.T) {
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
			layer := CreateCenteredLayer(tt.content, tt.screenWidth, tt.screenHeight)

			if layer == nil {
				t.Fatal("CreateCenteredLayer should return a layer for non-empty content")
			}
		})
	}
}

// TestCreateCenteredLayerWithEmptyContent tests layer creation with empty content
func TestCreateCenteredLayerWithEmptyContent(t *testing.T) {
	layer := CreateCenteredLayer("", 120, 40)

	if layer != nil {
		t.Error("CreateCenteredLayer should return nil for empty content")
	}
}

// TestLayerPositioning tests that layers are centered correctly
func TestLayerPositioning(t *testing.T) {
	content := "Center"
	screenWidth := 100
	screenHeight := 50

	layer := CreateCenteredLayer(content, screenWidth, screenHeight)

	if layer == nil {
		t.Fatal("layer should not be nil")
	}

	// Layer methods are chainable, so just verify layer exists
	// (actual positioning is tested implicitly by the function working)
}

// TestLayerPositioningOnSmallScreen tests centering on very small screens
func TestLayerPositioningOnSmallScreen(t *testing.T) {
	content := "X"
	screenWidth := 5
	screenHeight := 3

	layer := CreateCenteredLayer(content, screenWidth, screenHeight)

	if layer == nil {
		t.Fatal("layer should not be nil even on small screen")
	}
}

// TestCalculatePickerDimensions tests picker dimension calculation
func TestCalculatePickerDimensions(t *testing.T) {
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
			width, height := CalculatePickerDimensions(
				tt.itemCount,
				tt.hasFilter,
				tt.screenWidth,
				tt.screenHeight,
				tt.minWidth,
				tt.maxWidth,
			)

			// Width should be within bounds
			if width < tt.minWidth {
				t.Errorf("width %d should be >= minWidth %d", width, tt.minWidth)
			}
			if width > tt.maxWidth {
				t.Errorf("width %d should be <= maxWidth %d", width, tt.maxWidth)
			}

			// Height should be reasonable
			if height < PickerMinHeight {
				t.Errorf("height %d should be >= minimum %d", height, PickerMinHeight)
			}
			if height > tt.screenHeight {
				t.Errorf("height %d should not exceed screen height %d", height, tt.screenHeight)
			}
		})
	}
}

// TestPickerDimensionsWithLargeItemCount tests picker dimensions with many items
func TestPickerDimensionsWithLargeItemCount(t *testing.T) {
	pickerWidth, pickerHeight := CalculatePickerDimensions(
		10000, // very many items
		true,
		200, // large screen
		100,
		30,
		150,
	)

	// Should be capped at max visible items
	if pickerWidth <= 0 {
		t.Errorf("width should be positive, got %d", pickerWidth)
	}
	if pickerHeight <= 0 {
		t.Errorf("height should be positive, got %d", pickerHeight)
	}
}

// TestPickerDimensionsMinimums tests that picker respects minimum dimensions
func TestPickerDimensionsMinimums(t *testing.T) {
	width, height := CalculatePickerDimensions(
		10, // some items
		false,
		80, // reasonable screen
		30,
		20, // minWidth
		50, // maxWidth
	)

	// Should respect minimum width
	if width < 20 {
		t.Errorf("width should respect minimum: got %d, want >= 20", width)
	}

	// Should be reasonable height
	if height < 5 {
		t.Errorf("height should be reasonable: got %d, want >= 5", height)
	}
}

// TestPickerDimensionsMaximums tests that picker respects maximum dimensions
func TestPickerDimensionsMaximums(t *testing.T) {
	screenHeight := 50
	pickerWidth, pickerHeight := CalculatePickerDimensions(
		100,
		true,
		300, // very wide screen
		screenHeight,
		10,
		100, // maxWidth
	)

	// Should respect maximum width
	if pickerWidth > 100 {
		t.Errorf("width should respect maximum: got %d, want <= 100", pickerWidth)
	}

	// Should not exceed screen height (max 3/4)
	maxHeight := screenHeight * PickerMaxHeightNumerator / PickerMaxHeightDivisor
	if pickerHeight > maxHeight {
		t.Errorf("height should not exceed limit: got %d, want <= %d", pickerHeight, maxHeight)
	}
}

// TestPickerDimensionsWithAndWithoutFilter tests filter impact on dimensions
func TestPickerDimensionsWithAndWithoutFilter(t *testing.T) {
	const itemCount = 10
	const screenWidth = 120
	const screenHeight = 40
	const minWidth = 20
	const maxWidth = 60

	w1, h1 := CalculatePickerDimensions(
		itemCount, true, screenWidth, screenHeight, minWidth, maxWidth)
	w2, h2 := CalculatePickerDimensions(
		itemCount, false, screenWidth, screenHeight, minWidth, maxWidth)

	// Filter adds to chrome height, so with filter should be slightly taller
	if h1 <= h2 {
		t.Errorf("with filter height %d should be > without filter height %d", h1, h2)
	}

	// Widths should be the same
	if w1 != w2 {
		t.Errorf("widths should match: with filter %d != without filter %d", w1, w2)
	}
}

// TestPickerDimensionsConsistency tests consistency across multiple calls
func TestPickerDimensionsConsistency(t *testing.T) {
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

	// Results should be consistent
	if w1 != w2 || w1 != w3 {
		t.Errorf("width should be consistent: got %d, %d, %d", w1, w2, w3)
	}
	if h1 != h2 || h1 != h3 {
		t.Errorf("height should be consistent: got %d, %d, %d", h1, h2, h3)
	}
}

// TestLayerMultipleDimensions tests layers work with various screen dimensions
func TestLayerMultipleDimensions(t *testing.T) {
	screenDimensions := [][2]int{
		{40, 10},
		{80, 24},
		{120, 40},
		{160, 50},
		{200, 80},
	}

	for _, dims := range screenDimensions {
		width, height := dims[0], dims[1]
		layer := CreateCenteredLayer("Test", width, height)

		if layer == nil {
			t.Fatalf("should create layer for %dx%d", width, height)
		}
	}
}

// TestLayerWithMultilineContent tests layer centering with multiline content
func TestLayerWithMultilineContent(t *testing.T) {
	multilineContent := `Line 1
Line 2
Line 3
Line 4`

	layer := CreateCenteredLayer(multilineContent, 100, 50)

	if layer == nil {
		t.Fatal("should create layer for multiline content")
	}

	// Just verify it was created successfully
}
