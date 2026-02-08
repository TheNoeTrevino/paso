package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCalculateViewportSize_ZeroWidth ensures viewport defaults to minimum (3) when terminal width is 0.
// Edge case: Terminal not fully initialized yet.
func TestCalculateViewportSize_ZeroWidth(t *testing.T) {
	state := NewUIState()
	state.SetWidth(0)

	got := state.ViewportSize()
	assert.Equal(t, 3, got, "ViewportSize() with width=0, want 3 (minimum)")
}

// TestCalculateViewportSize_NarrowTerminal ensures viewport is at least 1 even with very small width.
// Edge case: User has extremely narrow terminal (< column width).
func TestCalculateViewportSize_NarrowTerminal(t *testing.T) {
	state := NewUIState()

	// Set width smaller than one column (46 chars)
	state.SetWidth(20)

	got := state.ViewportSize()
	assert.GreaterOrEqual(t, got, 1, "ViewportSize() with width=20, want >= 1")
}

// TestScrollViewportLeft_AtBoundary ensures scroll left at offset 0 is a no-op.
// Edge case: User presses scroll-left when already at leftmost position.
func TestScrollViewportLeft_AtBoundary(t *testing.T) {
	state := NewUIState()
	state.ViewportOffset = 0

	scrolled := state.ScrollViewportLeft()

	assert.False(t, scrolled, "ScrollViewportLeft() at offset=0 returned true, want false")
	assert.Equal(t, 0, state.ViewportOffset)
}

// TestScrollViewportRight_AtBoundary ensures scroll right at last column is a no-op.
// Edge case: User presses scroll-right when viewport shows the last column.
func TestScrollViewportRight_AtBoundary(t *testing.T) {
	state := NewUIState()
	state.SetWidth(300)      // Large enough for 6 columns
	state.ViewportOffset = 2 // Offset at position 2

	// Total columns = 5, viewport size = 6, so offset=0 already shows all columns
	// With offset=2, trying to scroll right when (2 + 6) >= 5 should fail
	columnsLen := 5

	scrolled := state.ScrollViewportRight(columnsLen)

	// Should not scroll since offset(2) + viewportSize(6) = 8 >= columnsLen(5)
	assert.False(t, scrolled, "ScrollViewportRight() at boundary returned true, want false")
	assert.Equal(t, 2, state.ViewportOffset)
}

// TestAdjustViewportAfterColumnRemoval_EmptyColumns ensures viewport resets when all columns deleted.
// Edge case: User deletes the last remaining column.
func TestAdjustViewportAfterColumnRemoval_EmptyColumns(t *testing.T) {
	state := NewUIState()
	state.ViewportOffset = 3 // Offset at position 3

	// Adjust after all columns are deleted
	state.AdjustViewportAfterColumnRemoval(0, 0)

	assert.Equal(t, 0, state.ViewportOffset)
}

// TestEnsureSelectionVisible_SelectionBeyondViewport ensures viewport auto-scrolls to show selection.
// Edge case: User navigates to column outside current viewport.
func TestEnsureSelectionVisible_SelectionBeyondViewport(t *testing.T) {
	state := NewUIState()
	state.SetWidth(100) // Width gives 2 calculated columns, but minimum is 3
	// ViewportSize = max(3, (100 - 4) / 46) = max(3, 2) = 3 (minimum enforced)

	state.ViewportOffset = 0 // Show columns 0-2

	// Select column 3 (beyond viewport of size 3)
	state.EnsureSelectionVisible(3)

	// Viewport should adjust so column 3 is visible
	// New offset should be: 3 - viewportSize + 1 = 3 - 3 + 1 = 1
	expectedOffset := 1
	assert.Equal(t, expectedOffset, state.ViewportOffset)

	// Test left side: select column 0 when viewport is at offset 1
	state.EnsureSelectionVisible(0)
	assert.Equal(t, 0, state.ViewportOffset)
}
