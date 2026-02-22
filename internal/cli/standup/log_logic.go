package standup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/thenoetrevino/paso/internal/cli/styles"
	"github.com/thenoetrevino/paso/internal/config/colors"
)

const logContentTruncateLength = 80

// LogResult represents the output of a successful log operation
type LogResult struct {
	ID        int
	ProjectID int
	Content   string
	CreatedAt string
}

// FormatLogQuiet generates minimal output (log ID only)
func FormatLogQuiet(logID int) string {
	return fmt.Sprintf("%d\n", logID)
}

// FormatLogJSON generates the JSON output structure
func FormatLogJSON(result *LogResult) map[string]any {
	return map[string]any{
		"success": true,
		"log": map[string]any{
			"id":         result.ID,
			"project_id": result.ProjectID,
			"content":    result.Content,
			"created_at": result.CreatedAt,
		},
	}
}

// FormatLogHuman generates human-readable output
func FormatLogHuman(result *LogResult, colorScheme colors.ColorScheme) string {
	details := []styles.Detail{
		{Key: "ID", Value: fmt.Sprintf("%d", result.ID)},
		{Key: "Content", Value: styles.TruncateString(result.Content, logContentTruncateLength)},
		{Key: "Time", Value: result.CreatedAt},
	}
	return styles.RenderSuccessWithDetails("Standup logged", details, colorScheme)
}

// openEditor opens the user's preferred editor and returns the content written.
func openEditor(ctx context.Context) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	tmpFile, err := os.CreateTemp("", "paso-standup-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	// Write a hint comment
	_, _ = tmpFile.WriteString("\n# Enter your standup log above. Lines starting with # are ignored.\n")
	_ = tmpFile.Close()

	cmd := exec.CommandContext(ctx, editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read editor content: %w", err)
	}

	// Strip comment lines and trim
	var lines []string
	for line := range strings.SplitSeq(string(content), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			lines = append(lines, line)
		}
	}

	return strings.TrimSpace(strings.Join(lines, "\n")), nil
}
