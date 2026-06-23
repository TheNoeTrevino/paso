package task

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"
	"github.com/stretchr/testify/assert"

	"github.com/thenoetrevino/paso/internal/cli/styles"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestWriteIndentedLines(t *testing.T) {
	t.Parallel()

	t.Run("indents every source line", func(t *testing.T) {
		t.Parallel()
		var b strings.Builder
		WriteIndentedLines(&b, "alpha\nbeta", "  ", lipgloss.NewStyle())

		lines := strings.Split(strings.TrimRight(fixtures.StripANSI(b.String()), "\n"), "\n")
		assert.Equal(t, []string{"  alpha", "  beta"}, lines)
	})

	t.Run("wraps long lines to the same indent within the card width", func(t *testing.T) {
		t.Parallel()
		indent := "    "
		long := strings.Repeat("word ", 40)

		var b strings.Builder
		WriteIndentedLines(&b, long, indent, lipgloss.NewStyle())

		lines := strings.Split(strings.TrimRight(fixtures.StripANSI(b.String()), "\n"), "\n")
		assert.Greater(t, len(lines), 1, "long text should wrap")
		for _, l := range lines {
			// Continuation lines keep the indent instead of collapsing to the margin.
			assert.True(t, strings.HasPrefix(l, indent), "line %q lost its indent", l)
			// And stay within the card so the border render doesn't re-wrap them.
			// Measure visual columns (not bytes) since the wrap library and the
			// card width are both in terminal cells.
			assert.LessOrEqual(t, runewidth.StringWidth(l), styles.CardContentWidth(), "line %q exceeds card width", l)
		}
	})
}
