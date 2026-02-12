package styles

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/config/colors"
)

var updateGolden = flag.Bool("update", false, "update golden files")

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func assertGoldenOutput(t *testing.T, name, actual string) {
	t.Helper()
	goldenPath := filepath.Join("testdata", name+".golden")

	if *updateGolden {
		err := os.MkdirAll("testdata", 0o755)
		require.NoError(t, err)
		err = os.WriteFile(goldenPath, []byte(actual), 0o644)
		require.NoError(t, err)
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Logf("Golden file does not exist, creating: %s\nRun with -update to regenerate", goldenPath)
			err := os.MkdirAll("testdata", 0o755)
			require.NoError(t, err)
			err = os.WriteFile(goldenPath, []byte(actual), 0o644)
			require.NoError(t, err)
			return
		}
		t.Fatalf("golden file not found: %s (run with -update to create)", goldenPath)
	}

	assert.Equal(t, string(expected), actual, "golden file mismatch for %s\n\nRun with -update to regenerate", name)
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

func TestRenderSuccess_GoldenOutput(t *testing.T) {
	t.Parallel()

	messages := []struct {
		name    string
		message string
	}{
		{"simple", "Task 42 deleted successfully"},
		{"with-quotes", "Active assignee set to 'john'"},
		{"with-at", "Task 42 assigned to @alice"},
		{"estimate-set", "Task 42 estimate set to 2d"},
		{"estimate-cleared", "Task 42 estimate cleared"},
		{"db-added", "Database 'cloud' added successfully"},
		{"project-linked", `Linked project "MyApp" (id: 5) to branch "feature/auth"`},
		{"project-unlinked", `Unlinked branch "main" from project "MyApp" (id: 5)`},
		{"plugin-installed", "OpenCode plugin installed"},
		{"hook-registered", "Registered SessionStart hook"},
		{"hook-removed", "Removed SessionStart hook"},
	}

	for schemeName, scheme := range allColorSchemes() {
		for _, msg := range messages {
			name := schemeName + "-" + msg.name
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				output := RenderSuccess(msg.message, scheme)
				stripped := stripANSI(output)
				assertGoldenOutput(t, "success-"+name, stripped)
			})
		}
	}
}

func TestRenderSuccessWithDetails_GoldenOutput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		message string
		details []Detail
	}{
		{
			name:    "task-created",
			message: "Task created successfully",
			details: []Detail{
				{Key: "ID", Value: "42"},
				{Key: "Title", Value: "Fix authentication bug"},
				{Key: "Project", Value: "MyApp"},
				{Key: "Priority", Value: "high"},
				{Key: "Type", Value: "task"},
			},
		},
		{
			name:    "project-created",
			message: "Project created successfully",
			details: []Detail{
				{Key: "ID", Value: "5"},
				{Key: "Name", Value: "MyApp"},
				{Key: "Description", Value: "Main application project"},
				{Key: "Branch", Value: "feature/auth"},
			},
		},
		{
			name:    "label-created",
			message: "Label created successfully",
			details: []Detail{
				{Key: "ID", Value: "3"},
				{Key: "Name", Value: "bug"},
				{Key: "Color", Value: "#FF0000"},
			},
		},
		{
			name:    "claude-installed",
			message: "Claude Code integration installed",
			details: []Detail{
				{Key: "Settings", Value: "/home/user/.claude/settings.json"},
			},
		},
		{
			name:    "opencode-installed",
			message: "OpenCode plugin installed",
			details: []Detail{
				{Key: "Config", Value: "/home/user/.config/opencode/opencode.json"},
			},
		},
	}

	for schemeName, scheme := range allColorSchemes() {
		for _, tc := range cases {
			name := schemeName + "-" + tc.name
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				output := RenderSuccessWithDetails(tc.message, tc.details, scheme)
				stripped := stripANSI(output)
				assertGoldenOutput(t, "success-details-"+name, stripped)
			})
		}
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
