package task

import (
	"regexp"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"

	"github.com/thenoetrevino/paso/internal/cli/styles"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

func TestWriteIndentedLines(t *testing.T) {
	t.Parallel()

	t.Run("indents every source line", func(t *testing.T) {
		t.Parallel()
		var b strings.Builder
		WriteIndentedLines(&b, "alpha\nbeta", "  ", lipgloss.NewStyle())

		lines := strings.Split(strings.TrimRight(stripANSI(b.String()), "\n"), "\n")
		assert.Equal(t, []string{"  alpha", "  beta"}, lines)
	})

	t.Run("wraps long lines to the same indent within the card width", func(t *testing.T) {
		t.Parallel()
		indent := "    "
		long := strings.Repeat("word ", 40)

		var b strings.Builder
		WriteIndentedLines(&b, long, indent, lipgloss.NewStyle())

		lines := strings.Split(strings.TrimRight(stripANSI(b.String()), "\n"), "\n")
		assert.Greater(t, len(lines), 1, "long text should wrap")
		for _, l := range lines {
			// Continuation lines keep the indent instead of collapsing to the margin.
			assert.True(t, strings.HasPrefix(l, indent), "line %q lost its indent", l)
			// And stay within the card so the border render doesn't re-wrap them.
			assert.LessOrEqual(t, len(l), styles.CardContentWidth(), "line %q exceeds card width", l)
		}
	})
}
