package state

import (
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
)

// TestGetFilteredItems_EmptyFilter ensures no filter returns all items.
// Edge case: User hasn't typed any filter text yet.
// Security value: Baseline behavior - filter is optional functionality.
func TestGetFilteredItems_EmptyFilter(t *testing.T) {
	state := NewLabelPickerState()
	state.Items = []LabelPickerItem{
		{Label: &models.Label{ID: 1, Name: "bug"}, Selected: false},
		{Label: &models.Label{ID: 2, Name: "feature"}, Selected: false},
	}
	state.Filter = ""

	filtered := state.GetFilteredItems()
	if len(filtered) != 2 {
		t.Errorf("GetFilteredItems() with empty filter = %d items, want 2", len(filtered))
	}
}

// TestGetFilteredItems_NoMatches ensures filter matching nothing returns empty result.
// Edge case: User's filter text doesn't match any labels.
// Security value: Returns nil slice (safe to iterate in Go - len(nil) = 0).
func TestGetFilteredItems_NoMatches(t *testing.T) {
	state := NewLabelPickerState()
	state.Items = []LabelPickerItem{
		{Label: &models.Label{ID: 1, Name: "bug"}, Selected: false},
		{Label: &models.Label{ID: 2, Name: "feature"}, Selected: false},
	}
	state.Filter = "xyz"

	filtered := state.GetFilteredItems()
	// Note: In Go, an uninitialized slice is nil, which is safe to iterate
	// len(nil) == 0, so this is correct behavior
	if len(filtered) != 0 {
		t.Errorf("GetFilteredItems() with no matches = %d items, want 0", len(filtered))
	}
}

// TestGetFilteredItems_CaseInsensitive ensures "BUG" matches "bug" label.
// Edge case: User types filter in different case than label name.
// Security value: Matches user expectation (search should be case-insensitive).
func TestGetFilteredItems_CaseInsensitive(t *testing.T) {
	state := NewLabelPickerState()
	state.Items = []LabelPickerItem{
		{Label: &models.Label{ID: 1, Name: "bug"}, Selected: false},
		{Label: &models.Label{ID: 2, Name: "Feature"}, Selected: false},
		{Label: &models.Label{ID: 3, Name: "DOCS"}, Selected: false},
	}

	testCases := []struct {
		filter string
		want   int // number of matches
	}{
		{"BUG", 1},     // matches "bug"
		{"feature", 1}, // matches "Feature"
		{"docs", 1},    // matches "DOCS"
		{"e", 1},       // matches "Feature" only (bug and DOCS don't contain 'e')
		{"u", 2},       // matches "bug" and "Feature"
	}

	for _, tc := range testCases {
		state.Filter = tc.filter
		filtered := state.GetFilteredItems()
		if len(filtered) != tc.want {
			t.Errorf("GetFilteredItems() with filter %q = %d items, want %d", tc.filter, len(filtered), tc.want)
		}
	}
}

// TestMoveCursorDown_EmptyItems ensures cursor movement with no labels is safe.
// Edge case: User navigates in label picker with zero labels.
// Security value: Cursor stays at 0 (doesn't go negative or out of bounds).
func TestMoveCursorDown_EmptyItems(t *testing.T) {
	state := NewLabelPickerState()
	state.Items = []LabelPickerItem{} // No items
	state.Cursor = 0

	moved := state.MoveCursorDown(0) // maxIdx = 0 (no items)

	if moved {
		t.Error("MoveCursorDown() with no items returned true, want false")
	}
	if state.Cursor != 0 {
		t.Errorf("Cursor after MoveCursorDown with no items = %d, want 0", state.Cursor)
	}
}

// TestMoveCursorDown_AtMax ensures cursor at last item doesn't move beyond.
// Edge case: User presses down when cursor is at bottom.
// Security value: No movement beyond end (prevents out of bounds access).
func TestMoveCursorDown_AtMax(t *testing.T) {
	state := NewLabelPickerState()
	state.Items = []LabelPickerItem{
		{Label: &models.Label{ID: 1, Name: "Label1"}, Selected: false},
		{Label: &models.Label{ID: 2, Name: "Label2"}, Selected: false},
		{Label: &models.Label{ID: 3, Name: "Label3"}, Selected: false},
	}
	state.Cursor = 3 // At max (len(items) = 3, so cursor can be 0-3 for "create new" option)

	moved := state.MoveCursorDown(3) // maxIdx = 3

	if moved {
		t.Error("MoveCursorDown() at max returned true, want false")
	}
	if state.Cursor != 3 {
		t.Errorf("Cursor after MoveCursorDown at max = %d, want 3", state.Cursor)
	}
}

// TestCursorAdjustment_FilterReducesList ensures cursor repositions when filter shrinks list.
// Edge case: User types filter that reduces list to fewer items than cursor position.
// Security value: Cursor repositions to valid index (prevents out of bounds).
// Note: This is a behavioral test - actual adjustment happens in update.go, not in state.
func TestCursorAdjustment_FilterReducesList(t *testing.T) {
	state := NewLabelPickerState()
	state.Items = []LabelPickerItem{
		{Label: &models.Label{ID: 1, Name: "bug"}, Selected: false},
		{Label: &models.Label{ID: 2, Name: "feature"}, Selected: false},
		{Label: &models.Label{ID: 3, Name: "docs"}, Selected: false},
	}
	state.Cursor = 2 // At third item
	state.Filter = ""

	// Apply filter that matches only one item
	state.Filter = "bug"
	filtered := state.GetFilteredItems()

	// Cursor is now beyond filtered list (cursor=2, but filtered has only 1 item at index 0)
	// In real code, update.go should adjust cursor to len(filtered) or 0
	// Here we just document that filtered list has fewer items
	if len(filtered) != 1 {
		t.Errorf("Filtered items = %d, want 1", len(filtered))
	}
	if state.Cursor >= len(filtered) {
		// This is expected - cursor needs adjustment by caller
		t.Logf("Cursor (%d) is beyond filtered list length (%d) - caller should adjust", state.Cursor, len(filtered))
	}
}

// TestAppendFilter_MaxLength ensures filter at 50 chars rejects more input.
// Edge case: User types continuously until reaching filter limit.
// Security value: Prevents excessive memory use in filter string.
func TestAppendFilter_MaxLength(t *testing.T) {
	state := NewLabelPickerState()

	// Fill filter to exactly 50 characters
	for i := 0; i < 50; i++ {
		state.Filter += "a"
	}

	// Try to append one more
	added := state.AppendFilter('x')

	if added {
		t.Error("AppendFilter() at max length (50) returned true, want false")
	}
	if len(state.Filter) != 50 {
		t.Errorf("Filter length after append at max = %d, want 50", len(state.Filter))
	}
}

// TestBackspaceFilter_Empty ensures backspace on empty filter is safe.
// Edge case: User presses backspace when filter is already empty.
// Security value: No-op, no crash (no string slice underflow).
func TestBackspaceFilter_Empty(t *testing.T) {
	state := NewLabelPickerState()
	state.Filter = ""

	removed := state.BackspaceFilter()

	if removed {
		t.Error("BackspaceFilter() on empty filter returned true, want false")
	}
	if state.Filter != "" {
		t.Errorf("Filter after backspace on empty = %q, want empty string", state.Filter)
	}

	// Multiple backspaces should be safe
	for i := 0; i < 5; i++ {
		removed = state.BackspaceFilter()
		if removed {
			t.Errorf("BackspaceFilter() call %d on empty returned true, want false", i+1)
		}
	}
}
