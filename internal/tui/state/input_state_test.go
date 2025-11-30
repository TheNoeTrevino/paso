package state

import (
	"strings"
	"testing"
)

// TestAppendChar_MaxLength ensures buffer at 100 chars rejects more input.
// Edge case: User types continuously until reaching buffer limit.
// Security value: Prevents buffer overflow (unbounded memory growth).
func TestAppendChar_MaxLength(t *testing.T) {
	state := NewInputState()

	// Fill buffer to exactly 100 characters
	state.Buffer = strings.Repeat("a", 100)

	// Try to append one more character
	added := state.AppendChar('x')

	if added {
		t.Error("AppendChar() at max length (100) returned true, want false")
	}
	if len(state.Buffer) != 100 {
		t.Errorf("Buffer length after append at max = %d, want 100", len(state.Buffer))
	}
	if strings.Contains(state.Buffer, "x") {
		t.Error("AppendChar() at max length modified buffer, want no change")
	}
}

// TestAppendChar_AtMaxLength ensures exactly at limit, one more char is rejected.
// Edge case: Boundary condition at exactly maxLength.
// Security value: Validates buffer overflow protection at exact boundary.
func TestAppendChar_AtMaxLength(t *testing.T) {
	state := NewInputState()

	// Add exactly 100 characters
	for i := 0; i < 100; i++ {
		added := state.AppendChar('a')
		if !added {
			t.Fatalf("AppendChar() failed at character %d, want success until 100", i+1)
		}
	}

	// Verify we can't add more
	added := state.AppendChar('b')
	if added {
		t.Error("AppendChar() at position 101 returned true, want false")
	}

	// Verify length is still 100
	if len(state.Buffer) != 100 {
		t.Errorf("Buffer length = %d, want 100", len(state.Buffer))
	}
}

// TestBackspace_EmptyBuffer ensures backspace on empty string is safe.
// Edge case: User presses backspace repeatedly when buffer is empty.
// Security value: Prevents string slice underflow.
func TestBackspace_EmptyBuffer(t *testing.T) {
	state := NewInputState()
	state.Buffer = ""

	// Try backspace on empty buffer
	removed := state.Backspace()

	if removed {
		t.Error("Backspace() on empty buffer returned true, want false")
	}
	if state.Buffer != "" {
		t.Errorf("Buffer after backspace on empty = %q, want empty string", state.Buffer)
	}

	// Try multiple backspaces to ensure stability
	for i := 0; i < 5; i++ {
		removed = state.Backspace()
		if removed {
			t.Errorf("Backspace() call %d on empty buffer returned true, want false", i+1)
		}
	}
}

// TestIsEmpty_WhitespaceOnly ensures detection of whitespace-only input.
// Edge case: User enters only spaces/tabs, then submits.
// Security value: Prevents empty column names in database.
func TestIsEmpty_WhitespaceOnly(t *testing.T) {
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
			state := NewInputState()
			state.Buffer = tc.buffer

			got := state.IsEmpty()
			if got != tc.want {
				t.Errorf("IsEmpty() with buffer %q = %v, want %v", tc.buffer, got, tc.want)
			}
		})
	}
}

// TestTrimmedBuffer_LeadingTrailingSpaces ensures input sanitization works.
// Edge case: User enters text with leading/trailing whitespace.
// Security value: Clean data for database storage (no accidental whitespace in column names).
func TestTrimmedBuffer_LeadingTrailingSpaces(t *testing.T) {
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
			state := NewInputState()
			state.Buffer = tc.buffer

			got := state.TrimmedBuffer()
			if got != tc.want {
				t.Errorf("TrimmedBuffer() with buffer %q = %q, want %q", tc.buffer, got, tc.want)
			}
		})
	}
}
