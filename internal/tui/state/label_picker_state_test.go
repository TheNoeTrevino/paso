package state

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/models"
)

// TestGetFilteredItems_EmptyFilter ensures no filter returns all items.
// Edge case: User hasn't typed any filter text yet.
func TestGetFilteredItems_EmptyFilter(t *testing.T) {
	t.Parallel()
	state := NewLabelPickerState()
	state.Items = []LabelPickerItem{
		{Label: &models.Label{ID: 1, Name: "bug"}, Selected: false},
		{Label: &models.Label{ID: 2, Name: "feature"}, Selected: false},
	}
	state.Filter = ""

	filtered := state.GetFilteredItems()
	assert.Len(t, filtered, 2)
}

// TestGetFilteredItems_NoMatches ensures filter matching nothing returns empty result.
// Edge case: User's filter text doesn't match any labels.
func TestGetFilteredItems_NoMatches(t *testing.T) {
	t.Parallel()
	state := NewLabelPickerState()
	state.Items = []LabelPickerItem{
		{Label: &models.Label{ID: 1, Name: "bug"}, Selected: false},
		{Label: &models.Label{ID: 2, Name: "feature"}, Selected: false},
	}
	state.Filter = "xyz"

	filtered := state.GetFilteredItems()
	// Note: In Go, an uninitialized slice is nil, which is safe to iterate
	// len(nil) == 0, so this is correct behavior
	assert.Len(t, filtered, 0)
}

// TestGetFilteredItems_CaseInsensitive ensures "BUG" matches "bug" label.
// Edge case: User types filter in different case than label name.
func TestGetFilteredItems_CaseInsensitive(t *testing.T) {
	t.Parallel()
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
		assert.Len(t, filtered, tc.want, "GetFilteredItems() with filter %q", tc.filter)
	}
}

// TestMoveCursorDown_EmptyItems ensures cursor movement with no labels is safe.
// Edge case: User navigates in label picker with zero labels.
func TestMoveCursorDown_EmptyItems(t *testing.T) {
	t.Parallel()
	state := NewLabelPickerState()
	state.Items = []LabelPickerItem{} // No items
	state.Cursor = 0

	moved := state.MoveCursorDown(0) // maxIdx = 0 (no items)

	assert.False(t, moved, "MoveCursorDown() with no items returned true, want false")
	assert.Equal(t, 0, state.Cursor)
}

// TestMoveCursorDown_AtMax ensures cursor at last item doesn't move beyond.
// Edge case: User presses down when cursor is at bottom.
func TestMoveCursorDown_AtMax(t *testing.T) {
	t.Parallel()
	state := NewLabelPickerState()
	state.Items = []LabelPickerItem{
		{Label: &models.Label{ID: 1, Name: "Label1"}, Selected: false},
		{Label: &models.Label{ID: 2, Name: "Label2"}, Selected: false},
		{Label: &models.Label{ID: 3, Name: "Label3"}, Selected: false},
	}
	state.Cursor = 3 // At max (len(items) = 3, so cursor can be 0-3 for "create new" option)

	moved := state.MoveCursorDown(3) // maxIdx = 3

	assert.False(t, moved, "MoveCursorDown() at max returned true, want false")
	assert.Equal(t, 3, state.Cursor)
}

// TestCursorAdjustment_FilterReducesList ensures cursor repositions when filter shrinks list.
// Edge case: User types filter that reduces list to fewer items than cursor position.
// Note: This is a behavioral test - actual adjustment happens in update.go, not in state.
func TestCursorAdjustment_FilterReducesList(t *testing.T) {
	t.Parallel()
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
	assert.Len(t, filtered, 1)
	if state.Cursor >= len(filtered) {
		// This is expected - cursor needs adjustment by caller
		t.Logf("Cursor (%d) is beyond filtered list length (%d) - caller should adjust", state.Cursor, len(filtered))
	}
}

// TestAppendFilter_MaxLength ensures filter at 50 chars rejects more input.
// Edge case: User types continuously until reaching filter limit.
func TestAppendFilter_MaxLength(t *testing.T) {
	t.Parallel()
	state := NewLabelPickerState()

	// Fill filter to exactly 50 characters
	state.Filter = strings.Repeat("a", 50)

	// Try to append one more
	added := state.AppendFilter('x')

	assert.False(t, added, "AppendFilter() at max length (50) returned true, want false")
	assert.Equal(t, 50, len(state.Filter))
}

// TestBackspaceFilter_Empty ensures backspace on empty filter is safe.
// Edge case: User presses backspace when filter is already empty.
func TestBackspaceFilter_Empty(t *testing.T) {
	t.Parallel()
	state := NewLabelPickerState()
	state.Filter = ""

	removed := state.BackspaceFilter()

	assert.False(t, removed, "BackspaceFilter() on empty filter returned true, want false")
	assert.Equal(t, "", state.Filter)

	// Multiple backspaces should be safe
	for i := range 5 {
		removed = state.BackspaceFilter()
		assert.False(t, removed, "BackspaceFilter() call %d on empty returned true, want false", i+1)
	}
}
