package cli_test

import (
	"regexp"
	"testing"

	"github.com/thenoetrevino/paso/internal/cli/styles"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/testing/snapshots"
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
				output := styles.RenderSuccess(msg.message, scheme)
				stripped := stripANSI(output)
				helper := snapshots.NewHelper(t, "testdata/styles")
				helper.Compare("success-"+name, stripped)
			})
		}
	}
}

func TestRenderSuccessWithDetails_GoldenOutput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		message string
		details []styles.Detail
	}{
		{
			name:    "task-created",
			message: "Task created successfully",
			details: []styles.Detail{
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
			details: []styles.Detail{
				{Key: "ID", Value: "5"},
				{Key: "Name", Value: "MyApp"},
				{Key: "Description", Value: "Main application project"},
				{Key: "Branch", Value: "feature/auth"},
			},
		},
		{
			name:    "label-created",
			message: "Label created successfully",
			details: []styles.Detail{
				{Key: "ID", Value: "3"},
				{Key: "Name", Value: "bug"},
				{Key: "Color", Value: "#FF0000"},
			},
		},
		{
			name:    "claude-installed",
			message: "Claude Code integration installed",
			details: []styles.Detail{
				{Key: "Settings", Value: "/home/user/.claude/settings.json"},
			},
		},
		{
			name:    "opencode-installed",
			message: "OpenCode plugin installed",
			details: []styles.Detail{
				{Key: "Config", Value: "/home/user/.config/opencode/opencode.json"},
			},
		},
	}

	for schemeName, scheme := range allColorSchemes() {
		for _, tc := range cases {
			name := schemeName + "-" + tc.name
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				output := styles.RenderSuccessWithDetails(tc.message, tc.details, scheme)
				stripped := stripANSI(output)
				helper := snapshots.NewHelper(t, "testdata/styles")
				helper.Compare("success-details-"+name, stripped)
			})
		}
	}
}
