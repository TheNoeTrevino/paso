package state

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAppendChar_MaxLength ensures buffer at 100 chars rejects more input.
// Edge case: User types continuously until reaching buffer limit.
func TestAppendChar_MaxLength(t *testing.T) {
	t.Parallel()
	state := NewInputState()

	// Fill buffer to exactly 100 characters
	state.Buffer = strings.Repeat("a", 100)

	// Try to append one more character
	added := state.AppendChar('x')

	assert.False(t, added, "AppendChar() at max length (100) returned true, want false")
	assert.Equal(t, 100, len(state.Buffer))
	assert.False(t, strings.Contains(state.Buffer, "x"), "AppendChar() at max length modified buffer, want no change")
}

// TestAppendChar_AtMaxLength ensures exactly at limit, one more char is rejected.
// Edge case: Boundary condition at exactly maxLength.
func TestAppendChar_AtMaxLength(t *testing.T) {
	t.Parallel()
	state := NewInputState()

	// Add exactly 100 characters
	for i := 0; i < 100; i++ {
		added := state.AppendChar('a')
		require.True(t, added, "AppendChar() failed at character %d, want success until 100", i+1)
	}

	// Verify we can't add more
	added := state.AppendChar('b')
	assert.False(t, added, "AppendChar() at position 101 returned true, want false")

	// Verify length is still 100
	assert.Equal(t, 100, len(state.Buffer))
}

// TestBackspace_EmptyBuffer ensures backspace on empty string is safe.
// Edge case: User presses backspace repeatedly when buffer is empty.
func TestBackspace_EmptyBuffer(t *testing.T) {
	t.Parallel()
	state := NewInputState()
	state.Buffer = ""

	// Try backspace on empty buffer
	removed := state.Backspace()

	assert.False(t, removed, "Backspace() on empty buffer returned true, want false")
	assert.Equal(t, "", state.Buffer)

	// Try multiple backspaces to ensure stability
	for i := 0; i < 5; i++ {
		removed = state.Backspace()
		assert.False(t, removed, "Backspace() call %d on empty buffer returned true, want false", i+1)
	}
}

// TestIsEmpty_WhitespaceOnly ensures detection of whitespace-only input.
// Edge case: User enters only spaces/tabs, then submits.
func TestIsEmpty_WhitespaceOnly(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		buffer string
		want   bool
	}{
		{"Empty string", "", true},
		{"Single space", " ", true},
		{"Multiple spaces", "   ", true},
		{"Tabs", "\t\t", true},
		{"Mixed whitespace", " \t \n ", true},
		{"Valid text", "Todo", false},
		{"Text with spaces", "  Todo  ", false},
		{"Single char", "a", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()
			state := NewInputState()
			state.Buffer = tc.buffer

			got := state.IsEmpty()
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestTrimmedBuffer_LeadingTrailingSpaces ensures input sanitization works.
// Edge case: User enters text with leading/trailing whitespace.
func TestTrimmedBuffer_LeadingTrailingSpaces(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		buffer string
		want   string
	}{
		{"No whitespace", "Todo", "Todo"},
		{"Leading spaces", "  Todo", "Todo"},
		{"Trailing spaces", "Todo  ", "Todo"},
		{"Both sides", "  Todo  ", "Todo"},
		{"Tabs and spaces", "\t  Todo \t ", "Todo"},
		{"Internal spaces preserved", "In Progress", "In Progress"},
		{"Empty string", "", ""},
		{"Only spaces", "   ", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()
			state := NewInputState()
			state.Buffer = tc.buffer

			got := state.TrimmedBuffer()
			assert.Equal(t, tc.want, got)
		})
	}
}
