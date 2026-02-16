package styles

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/config/colors"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func allColorSchemes() map[string]colors.ColorScheme {
	return map[string]colors.ColorScheme{
		"default":    *colors.Default(),
		"monochrome": *colors.Monochrome(),
		"wave":       *colors.Wave(),
		"dragon":     *colors.Dragon(),
		"lotus":      *colors.Lotus(),
	}
}

func TestRenderSuccess_ContainsCheckmark(t *testing.T) {
	t.Parallel()

	for schemeName, scheme := range allColorSchemes() {
		t.Run(schemeName, func(t *testing.T) {
			t.Parallel()
			output := RenderSuccess("Test message", scheme)
			stripped := stripANSI(output)
			assert.Contains(t, stripped, "\u2713", "output should contain checkmark character")
			assert.Contains(t, stripped, "Test message", "output should contain the message")
			assert.True(t, stripped[len(stripped)-1] == '\n', "output should end with newline")
		})
	}
}

func TestRenderSuccessWithDetails_ContainsCheckmark(t *testing.T) {
	t.Parallel()

	details := []Detail{
		{Key: "ID", Value: "1"},
		{Key: "Name", Value: "test"},
	}

	for schemeName, scheme := range allColorSchemes() {
		t.Run(schemeName, func(t *testing.T) {
			t.Parallel()
			output := RenderSuccessWithDetails("Created successfully", details, scheme)
			stripped := stripANSI(output)
			assert.Contains(t, stripped, "\u2713", "output should contain checkmark character")
			assert.Contains(t, stripped, "Created successfully", "output should contain the message")
			assert.Contains(t, stripped, "ID: 1", "output should contain detail key-value pair")
			assert.Contains(t, stripped, "Name: test", "output should contain detail key-value pair")
		})
	}
}

func TestRenderSuccess_StructuralConsistency(t *testing.T) {
	t.Parallel()

	schemes := allColorSchemes()
	message := "Task 42 updated successfully"

	var outputs []string
	for _, scheme := range schemes {
		output := RenderSuccess(message, scheme)
		stripped := stripANSI(output)
		outputs = append(outputs, stripped)
	}

	// All stripped outputs should be identical regardless of color scheme
	for i := 1; i < len(outputs); i++ {
		assert.Equal(t, outputs[0], outputs[i],
			"stripped output should be identical across all color schemes")
	}
}

func TestRenderSuccessWithDetails_StructuralConsistency(t *testing.T) {
	t.Parallel()

	schemes := allColorSchemes()
	message := "Task created successfully"
	details := []Detail{
		{Key: "ID", Value: "42"},
		{Key: "Title", Value: "Fix bug"},
	}

	var outputs []string
	for _, scheme := range schemes {
		output := RenderSuccessWithDetails(message, details, scheme)
		stripped := stripANSI(output)
		outputs = append(outputs, stripped)
	}

	// All stripped outputs should be identical regardless of color scheme
	for i := 1; i < len(outputs); i++ {
		assert.Equal(t, outputs[0], outputs[i],
			"stripped output should be identical across all color schemes")
	}
}

func TestRenderError_ContainsXMark(t *testing.T) {
	t.Parallel()

	for schemeName, scheme := range allColorSchemes() {
		t.Run(schemeName, func(t *testing.T) {
			t.Parallel()
			output := RenderError("something went wrong", scheme)
			stripped := stripANSI(output)
			assert.Contains(t, stripped, "\u2717", "output should contain X mark character")
			assert.Contains(t, stripped, "Error:", "output should contain 'Error:' prefix")
			assert.Contains(t, stripped, "something went wrong", "output should contain the message")
		})
	}
}

func TestRenderError_StructuralConsistency(t *testing.T) {
	t.Parallel()

	schemes := allColorSchemes()
	message := "task 42 not found"

	var outputs []string
	for _, scheme := range schemes {
		output := RenderError(message, scheme)
		stripped := stripANSI(output)
		outputs = append(outputs, stripped)
	}

	for i := 1; i < len(outputs); i++ {
		assert.Equal(t, outputs[0], outputs[i],
			"stripped output should be identical across all color schemes")
	}
}
