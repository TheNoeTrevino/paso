package jira

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/thenoetrevino/paso/internal/cli/styles"
	"github.com/thenoetrevino/paso/internal/config/colors"
)

// issueKeyPattern matches Jira issue keys like PROJ-123, AB-1, MY_PROJ-42.
var issueKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]+-\d+$`)

// ValidateIssueKey validates a Jira issue key string (e.g., "PROJ-123").
func ValidateIssueKey(s string) (string, error) {
	if !issueKeyPattern.MatchString(s) {
		return "", fmt.Errorf("invalid issue key %q: must match format PROJ-123 (uppercase project prefix, hyphen, number)", s)
	}
	return s, nil
}

// importResult represents the result of importing a Jira issue.
type importResult struct {
	TaskID        int
	Title         string
	Description   string
	Project       string
	IssueKey      string
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
		{Key: "Jira Issue", Value: r.IssueKey},
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

	return styles.RenderSuccessWithDetails("Jira issue imported", details, colorScheme)
}
