package styles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDimColor(t *testing.T) {
	t.Parallel()
	t.Run("full dim returns black", func(t *testing.T) {
		result := DimColor("#FFFFFF", 1.0)
		assert.Equal(t, "#000000", result)
	})

	t.Run("no dim returns original", func(t *testing.T) {
		result := DimColor("#FF0000", 0.0)
		assert.Equal(t, "#FF0000", result)
	})

	t.Run("half dim reduces values", func(t *testing.T) {
		result := DimColor("#FF0000", 0.5)
		// 255 * 0.5 = 127 = 0x7F
		assert.Equal(t, "#7F0000", result)
	})

	t.Run("clamps intensity above 1", func(t *testing.T) {
		result := DimColor("#FFFFFF", 2.0)
		assert.Equal(t, "#000000", result)
	})

	t.Run("clamps intensity below 0", func(t *testing.T) {
		result := DimColor("#FF0000", -1.0)
		assert.Equal(t, "#FF0000", result)
	})

	t.Run("invalid hex returns original", func(t *testing.T) {
		result := DimColor("not-a-color", 0.5)
		assert.Equal(t, "not-a-color", result)
	})

	t.Run("3-char hex", func(t *testing.T) {
		result := DimColor("#FFF", 1.0)
		assert.Equal(t, "#000000", result)
	})

	t.Run("dim all channels", func(t *testing.T) {
		result := DimColor("#AABBCC", 0.5)
		// AA=170, BB=187, CC=204
		// 170*0.5=85=0x55, 187*0.5=93=0x5D, 204*0.5=102=0x66
		assert.Equal(t, "#555D66", result)
	})
}

func TestParseHex(t *testing.T) {
	t.Parallel()
	t.Run("parses 6-char hex", func(t *testing.T) {
		r, g, b, ok := parseHex("#FF0000")
		assert.True(t, ok)
		assert.Equal(t, uint8(255), r)
		assert.Equal(t, uint8(0), g)
		assert.Equal(t, uint8(0), b)
	})

	t.Run("parses 6-char hex without hash", func(t *testing.T) {
		r, g, b, ok := parseHex("00FF00")
		assert.True(t, ok)
		assert.Equal(t, uint8(0), r)
		assert.Equal(t, uint8(255), g)
		assert.Equal(t, uint8(0), b)
	})

	t.Run("parses 3-char hex", func(t *testing.T) {
		r, g, b, ok := parseHex("#F00")
		assert.True(t, ok)
		assert.Equal(t, uint8(255), r)
		assert.Equal(t, uint8(0), g)
		assert.Equal(t, uint8(0), b)
	})

	t.Run("parses 3-char hex FFF", func(t *testing.T) {
		r, g, b, ok := parseHex("#FFF")
		assert.True(t, ok)
		assert.Equal(t, uint8(255), r)
		assert.Equal(t, uint8(255), g)
		assert.Equal(t, uint8(255), b)
	})

	t.Run("returns false for invalid hex", func(t *testing.T) {
		_, _, _, ok := parseHex("xyz")
		assert.False(t, ok)
	})

	t.Run("returns false for wrong length", func(t *testing.T) {
		_, _, _, ok := parseHex("#FFFF")
		assert.False(t, ok)
	})

	t.Run("returns false for empty string", func(t *testing.T) {
		_, _, _, ok := parseHex("")
		assert.False(t, ok)
	})
}
