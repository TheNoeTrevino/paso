package state

import (
	"testing"
)

// TestCalculateViewportSize_ZeroWidth ensures viewport defaults to 1 when terminal width is 0.
// Edge case: Terminal not fully initialized yet.
// Security value: Prevents division by zero or negative viewport size.
func TestCalculateViewportSize_ZeroWidth(t *testing.T) {
	state := NewUIState()
	state.SetWidth(0)

	got := state.ViewportSize()
	if got != 1 {
		t.Errorf("ViewportSize() with width=0 = %d, want 1", got)
	}
}

// TestCalculateViewportSize_NarrowTerminal ensures viewport is at least 1 even with very small width.
// Edge case: User has extremely narrow terminal (< column width).
// Security value: Ensures minimum viewport of 1 column (prevents zero-column state).
func TestCalculateViewportSize_NarrowTerminal(t *testing.T) {
	state := NewUIState()

	// Set width smaller than one column (46 chars)
	state.SetWidth(20)

	got := state.ViewportSize()
	if got < 1 {
		t.Errorf("ViewportSize() with width=20 = %d, want >= 1", got)
	}
}

// TestScrollViewportLeft_AtBoundary ensures scroll left at offset 0 is a no-op.
// Edge case: User presses scroll-left when already at leftmost position.
// Security value: Prevents negative offset (array underflow).
func TestScrollViewportLeft_AtBoundary(t *testing.T) {
	state := NewUIState()
	state.SetViewportOffset(0)

	scrolled := state.ScrollViewportLeft()

	if scrolled {
		t.Error("ScrollViewportLeft() at offset=0 returned true, want false")
	}
	if state.ViewportOffset() != 0 {
		t.Errorf("ViewportOffset after scroll = %d, want 0", state.ViewportOffset())
	}
}

// TestScrollViewportRight_AtBoundary ensures scroll right at last column is a no-op.
// Edge case: User presses scroll-right when viewport shows the last column.
// Security value: Prevents offset beyond column count.
func TestScrollViewportRight_AtBoundary(t *testing.T) {
	state := NewUIState()
	state.SetWidth(300)        // Large enough for 6 columns
	state.SetViewportOffset(2) // Offset at position 2

	// Total columns = 5, viewport size = 6, so offset=0 already shows all columns
	// With offset=2, trying to scroll right when (2 + 6) >= 5 should fail
	columnsLen := 5

	scrolled := state.ScrollViewportRight(columnsLen)

	// Should not scroll since offset(2) + viewportSize(6) = 8 >= columnsLen(5)
	if scrolled {
		t.Error("ScrollViewportRight() at boundary returned true, want false")
	}
	if state.ViewportOffset() != 2 {
		t.Errorf("ViewportOffset after scroll = %d, want 2 (unchanged)", state.ViewportOffset())
	}
}

// TestAdjustViewportAfterColumnRemoval_EmptyColumns ensures viewport resets when all columns deleted.
// Edge case: User deletes the last remaining column.
// Security value: Prevents panic on empty state.
func TestAdjustViewportAfterColumnRemoval_EmptyColumns(t *testing.T) {
	state := NewUIState()
	state.SetViewportOffset(3) // Offset at position 3

	// Adjust after all columns are deleted
	state.AdjustViewportAfterColumnRemoval(0, 0)

	if state.ViewportOffset() != 0 {
		t.Errorf("ViewportOffset after removing all columns = %d, want 0", state.ViewportOffset())
	}
}

// TestEnsureSelectionVisible_SelectionBeyondViewport ensures viewport auto-scrolls to show selection.
// Edge case: User navigates to column outside current viewport.
// Security value: Ensures selection always accessible (prevents invisible selection state).
func TestEnsureSelectionVisible_SelectionBeyondViewport(t *testing.T) {
	state := NewUIState()
	state.SetWidth(100) // Enough for 2 columns (46 chars per column + 4 reserved)
	// ViewportSize should be 2: (100 - 4) / 46 = 2

	state.SetViewportOffset(0) // Show columns 0-1

	// Select column 3 (beyond viewport)
	state.EnsureSelectionVisible(3)

	// Viewport should adjust so column 3 is visible
	// New offset should be: 3 - viewportSize + 1 = 3 - 2 + 1 = 2
	expectedOffset := 2
	if state.ViewportOffset() != expectedOffset {
		t.Errorf("ViewportOffset after EnsureSelectionVisible(3) = %d, want %d", state.ViewportOffset(), expectedOffset)
	}

	// Test left side: select column 0 when viewport is at offset 2
	state.EnsureSelectionVisible(0)
	if state.ViewportOffset() != 0 {
		t.Errorf("ViewportOffset after EnsureSelectionVisible(0) from offset=2 = %d, want 0", state.ViewportOffset())
	}
}
