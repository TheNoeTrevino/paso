package gh

import (
	"fmt"
	"strconv"

	"github.com/thenoetrevino/paso/internal/cli/styles"
	"github.com/thenoetrevino/paso/internal/config/colors"
)

// ValidateIssueNumber parses and validates a GitHub issue number string.
func ValidateIssueNumber(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid issue number %q: must be a positive integer", s)
	}
	if n <= 0 {
		return 0, fmt.Errorf("invalid issue number %d: must be a positive integer", n)
	}
	return n, nil
}

// importResult represents the result of importing a GitHub issue.
type importResult struct {
	TaskID        int
	Title         string
	Description   string
	Project       string
	IssueNumber   int
	CommentsCount int
}

// GetID implements the quiet mode interface.
func (r *importResult) GetID() int {
	return r.TaskID
}

// PrettyPrint implements the PrettyPrintable interface for styled output.
func (r *importResult) PrettyPrint(colorScheme colors.ColorScheme) string {
	details := []styles.Detail{
		{Key: "Task ID", Value: strconv.Itoa(r.TaskID)},
		{Key: "GitHub Issue", Value: fmt.Sprintf("#%d", r.IssueNumber)},
		{Key: "Title", Value: r.Title},
		{Key: "Project", Value: r.Project},
	}

	if r.Description != "" {
		truncated := styles.TruncateString(r.Description, 60)
		details = append(details, styles.Detail{Key: "Description", Value: truncated})
	}

	details = append(details, styles.Detail{
		Key:   "Comments",
		Value: fmt.Sprintf("%d imported", r.CommentsCount),
	})

	return styles.RenderSuccessWithDetails("GitHub issue imported", details, colorScheme)
}
